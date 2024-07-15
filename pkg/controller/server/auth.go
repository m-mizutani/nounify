package server

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/interfaces"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
)

type middlewareFunc func(next http.Handler) http.Handler

func trimToken(token string) string {
	e := min(len(token), 8)
	return token[:e] + "..."
}

func authGitHubWebhook(secret string) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if not GitHub webhook
			if r.Header.Get("X-GitHub-Event") == "" {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := github.ValidatePayload(r, []byte(secret))
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			auth := model.NewGitHubAppAuth(r)
			r.Body = io.NopCloser(bytes.NewReader(payload))
			r = r.WithContext(ctxutil.WithGitHubAppAuth(r.Context(), auth))

			next.ServeHTTP(w, r)
		})
	}
}

func validateGitHubActionToken(authHdr string) (model.GitHubActionToken, error) {
	hdr := strings.SplitN(authHdr, " ", 2)

	// Skip if not Bearer token
	if len(hdr) != 2 || strings.ToLower(hdr[0]) != "bearer" {
		return nil, nil
	}

	jwksURL := "https://token.actions.githubusercontent.com/.well-known/jwks"

	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	token, err := jwt.ParseString(hdr[1], jwt.WithKeySet(set))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse JWT token as GitHub Action token").With("token", trimToken(hdr[1]))
	}

	claims, err := token.AsMap(context.Background())
	if err != nil {
		return nil, goerr.Wrap(err, "failed to convert JWT token to map").With("token", trimToken(hdr[1]))
	}

	return claims, nil
}

func authGitHubActionToken() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := validateGitHubActionToken(r.Header.Get("Authorization"))
			if claims == nil {
				if err != nil {
					ctxutil.Logger(r.Context()).Debug("failed to parse JWT token", "err", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ctxutil.WithGitHubActionToken(r.Context(), claims))
			next.ServeHTTP(w, r)
		})
	}
}

func validateGoogleIDToken(authHdr string) (map[string]any, error) {
	hdr := strings.SplitN(authHdr, " ", 2)

	// Skip if not Bearer token
	if len(hdr) != 2 || hdr[0] != "Bearer" {
		return nil, nil
	}

	jwksURL := "https://www.googleapis.com/oauth2/v3/certs"

	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	token, err := jwt.ParseString(hdr[1], jwt.WithKeySet(set))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse JWT token as Google ID Token").With("token", trimToken(hdr[1]))
	}

	claims, err := token.AsMap(context.Background())
	if err != nil {
		return nil, goerr.Wrap(err, "failed to convert JWT token to map").With("token", trimToken(hdr[1]))
	}

	return claims, nil
}

func authGoogleIDToken() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := validateGoogleIDToken(r.Header.Get("Authorization"))
			if claims == nil {
				if err != nil {
					ctxutil.Logger(r.Context()).Debug("failed to fetch JWK set", "err", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ctxutil.WithGoogleIDToken(r.Context(), claims))
			next.ServeHTTP(w, r)
		})
	}
}

type snsMessage struct {
	Type             string `json:"Type"`
	MessageId        string `json:"MessageId"`
	Token            string `json:"Token"`
	TopicArn         string `json:"TopicArn"`
	Subject          string `json:"Subject"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

func authAwsSNS() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth, err := validateSNSMessage(r)
			if err != nil {
				handleError(r.Context(), w, err)
				return
			}
			if auth != nil {
				r = r.WithContext(ctxutil.WithAwsSNSAuth(r.Context(), auth))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateSNSMessage(r *http.Request) (*model.AwsSNSAuth, error) {
	if r.Header.Get("X-Amz-Sns-Message-Id") == "" {
		return nil, nil
	}

	var msg snsMessage

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read request body")
	}
	defer r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body)) // refill the body

	// Unmarshal the JSON message
	if err = json.Unmarshal(body, &msg); err != nil {
		return nil, goerr.Wrap(err, "invalid JSON format").With("body", string(body))
	}

	cert, err := fetchAWSCert(r.Context(), msg.SigningCertURL)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to fetch certificate")
	}

	messageString, err := buildSNSMessageString(msg)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to build message string")
	}

	if err = verifyX509Signature(cert, msg.SignatureVersion, messageString, msg.Signature); err != nil {
		return nil, goerr.Wrap(err, "failed to verify signature")
	}

	return &model.AwsSNSAuth{
		Type:      msg.Type,
		MessageId: msg.MessageId,
		TopicArn:  msg.TopicArn,
		Timestamp: msg.Timestamp,
	}, nil
}

func fetchAWSCert(ctx context.Context, certURL string) (*x509.Certificate, error) {
	if u, err := url.Parse(certURL); err != nil {
		return nil, goerr.Wrap(err, "invalid URL").With("url", certURL)
	} else if u.Scheme != "https" || !strings.HasSuffix(u.Host, ".amazonaws.com") {
		return nil, goerr.New("unacceptable URL").With("url", certURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certURL, nil)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create request").With("url", certURL)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to fetch certificate").With("url", certURL)
	}
	defer resp.Body.Close()

	certData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read certificate data")
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, goerr.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse certificate")
	}

	return cert, nil
}

func buildSNSMessageString(message snsMessage) (string, error) {
	var msgParts []string

	switch message.Type {
	case "Notification":
		msgParts = []string{
			"Message", message.Message,
			"MessageId", message.MessageId,
		}

		if message.Subject != "" {
			msgParts = append(msgParts, "Subject", message.Subject)
		}

		msgParts = append(msgParts, []string{
			"Timestamp", message.Timestamp,
			"TopicArn", message.TopicArn,
			"Type", message.Type,
		}...)

	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		msgParts = []string{
			"Message", message.Message,
			"MessageId", message.MessageId,
			"SubscribeURL", message.SubscribeURL,
			"Timestamp", message.Timestamp,
			"Token", message.Token,
			"TopicArn", message.TopicArn,
			"Type", message.Type,
		}

	default:
		return "", errors.New("unknown message type")
	}

	return strings.Join(msgParts, "\n") + "\n", nil
}

func verifyX509Signature(cert *x509.Certificate, version, message, signature string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return goerr.Wrap(err, "failed to decode signature").With("signature", signature)
	}

	alg := map[string]x509.SignatureAlgorithm{
		"1": x509.SHA1WithRSA,
		"2": x509.SHA256WithRSA,
	}

	sigAlg, ok := alg[version]
	if !ok {
		return goerr.New("unsupported signature version").With("version", version)
	}

	if err := cert.CheckSignature(sigAlg, []byte(message), signatureBytes); err != nil {
		return goerr.Wrap(err, "failed to verify signature").
			With("message", message).
			With("signature", signature)
	}

	return nil
}

func authFromContext(ctx context.Context) model.AuthContext {
	var auth model.AuthContext

	if claims := ctxutil.GoogleIDToken(ctx); claims != nil {
		auth.Google = make(map[string]any, len(claims))
		for key, value := range claims {
			switch v := value.(type) {
			case time.Time:
				auth.Google[key] = v.Unix()
			default:
				auth.Google[key] = value
			}
		}
	}

	if claims := ctxutil.GitHubAppAuth(ctx); claims != nil {
		auth.GitHub.App = claims
	}

	if claims := ctxutil.GitHubActionToken(ctx); claims != nil {
		auth.GitHub.Action = make(map[string]any, len(claims))
		for key, value := range claims {
			switch v := value.(type) {
			case time.Time:
				auth.GitHub.Action[key] = v.Unix()
			default:
				auth.GitHub.Action[key] = value
			}
		}
	}

	if snsAuth := ctxutil.AwsSNSAuth(ctx); snsAuth != nil {
		auth.AWS.SNS = snsAuth
	}

	return auth
}

func authWithPolicy(policy interfaces.Policy, errCode int) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			input := model.AuthQueryInput{
				Method: r.Method,
				Path:   r.URL.Path,
				Header: map[string]string{},
			}

			for key := range r.Header {
				input.Header[key] = r.Header.Get(key)
			}

			ctx := r.Context()
			input.Auth = authFromContext(ctx)

			var output model.AuthQueryOutput
			if err := policy.Query(ctx, "data.auth", input, &output); err != nil {
				handleError(ctx, w, err)
				return
			}
			ctxutil.Logger(r.Context()).Debug("auth query result", "input", input, "output", output)

			if !output.Allow {
				handleError(ctx, w, types.ErrForbidden,
					handleErrorWithForceCode(errCode),
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

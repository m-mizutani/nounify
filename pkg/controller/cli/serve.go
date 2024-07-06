package cli

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/controller/server"
	"github.com/m-mizutani/nounify/pkg/usecase"
	"github.com/m-mizutani/opac"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v2"
)

func cmdServe() *cli.Command {
	var (
		addr        string
		slackToken  string
		policyFiles cli.StringSlice

		githubSecrets           cli.StringSlice
		enableGitHubActionToken bool
		enableGoogleIDToken     bool
		enableAuthErrOK         bool
	)

	return &cli.Command{
		Name:    "serve",
		Usage:   "Start HTTP server",
		Aliases: []string{"s"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "HTTP server address",
				Aliases:     []string{"a"},
				EnvVars:     []string{"NOUNIFY_ADDR"},
				Destination: &addr,
				Value:       "127.0.0.1:8080",
			},
			&cli.StringFlag{
				Name:        "slack-oauth-token",
				Usage:       "Slack OAuth token",
				EnvVars:     []string{"NOUNIFY_SLACK_OAUTH_TOKEN"},
				Destination: &slackToken,
				Required:    true,
			},
			&cli.StringSliceFlag{
				Name:        "policy",
				Usage:       "Policy path of file(s). When path is directory, all files in the directory are loaded. File extension must be .rego",
				Aliases:     []string{"p"},
				EnvVars:     []string{"NOUNIFY_POLICY"},
				Destination: &policyFiles,
				Required:    true,
			},

			&cli.StringSliceFlag{
				Name:        "github-secret",
				Usage:       "GitHub App webhook secret",
				EnvVars:     []string{"NOUNIFY_GITHUB_SECRET"},
				Destination: &githubSecrets,
			},
			&cli.BoolFlag{
				Name:        "github-action-token",
				Usage:       "Enable GitHub action token verification",
				EnvVars:     []string{"NOUNIFY_GITHUB_ACTION_TOKEN"},
				Destination: &enableGitHubActionToken,
			},
			&cli.BoolFlag{
				Name:        "google-id-token",
				Usage:       "Enable Google ID token verification",
				EnvVars:     []string{"NOUNIFY_GOOGLE_ID_TOKEN"},
				Destination: &enableGoogleIDToken,
			},
			&cli.BoolFlag{
				Name:        "auth-err-ok",
				Usage:       "Return 200 OK when authentication error",
				EnvVars:     []string{"NOUNIFY_AUTH_ERR_OK"},
				Destination: &enableAuthErrOK,
			},
		},

		Action: func(c *cli.Context) error {
			slackClient := slack.New(slackToken)
			policy, err := opac.New(opac.Files(policyFiles.Value()...))
			if err != nil {
				return goerr.Wrap(err, "failed to load policy files").With("files", policyFiles.Value())
			}

			uc := usecase.New(
				usecase.WithSlack(slackClient),
				usecase.WithPolicy(policy),
			)

			serverOptions := []server.Option{
				server.WithPolicy(policy),
			}
			for _, secret := range githubSecrets.Value() {
				serverOptions = append(serverOptions, server.WithGitHubSecret(secret))
			}
			if enableGoogleIDToken {
				serverOptions = append(serverOptions, server.WithGoogleIDTokenValidation())
			}
			if enableGitHubActionToken {
				serverOptions = append(serverOptions, server.WithGitHubActionTokenValidation())
			}

			s := &http.Server{
				Addr:              addr,
				ReadHeaderTimeout: 3 * time.Second,
				Handler:           server.New(uc, serverOptions...),
			}

			errCh := make(chan error, 1)

			go func() {
				if err := s.ListenAndServe(); err != nil {
					errCh <- goerr.Wrap(err, "failed to listen")
				}
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)

			select {
			case sig := <-sigCh:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := s.Shutdown(ctx); err != nil {
					return goerr.Wrap(err, "failed to shutdown server").With("signal", sig)
				}

			case err := <-errCh:
				return err
			}

			return nil
		},
	}
}

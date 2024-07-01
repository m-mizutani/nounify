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

		githubSecrets cli.StringSlice
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
				Name:        "policy-file",
				Usage:       "Policy file path",
				Aliases:     []string{"p"},
				EnvVars:     []string{"NOUNIFY_POLICY_FILE"},
				Destination: &policyFiles,
				Required:    true,
			},

			&cli.StringSliceFlag{
				Name:        "github-secret",
				Usage:       "GitHub App webhook secret",
				EnvVars:     []string{"NOUNIFY_GITHUB_SECRET"},
				Destination: &githubSecrets,
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

			s := &http.Server{
				Addr:              addr,
				ReadHeaderTimeout: 3 * time.Second,
				Handler:           server.New(uc),
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

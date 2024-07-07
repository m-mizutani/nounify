package config

import (
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/utils/logging"

	"github.com/urfave/cli/v2"
)

type Sentry struct {
	dsn string
	env string
}

func (x *Sentry) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "sentry-dsn",
			Usage:       "Sentry DSN for error reporting",
			EnvVars:     []string{"NOUNIFY_SENTRY_DSN"},
			Destination: &x.dsn,
		},
		&cli.StringFlag{
			Name:        "sentry-env",
			Usage:       "Sentry environment",
			EnvVars:     []string{"NOUNIFY_SENTRY_ENV"},
			Destination: &x.env,
		},
	}
}

func (x *Sentry) Configure() error {
	if x.dsn != "" {
		logging.Default().Info("Enable Sentry", "DSN", x.dsn, "env", x.env)
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         x.dsn,
			Environment: x.env,
		}); err != nil {
			return goerr.Wrap(err, "failed to initialize Sentry")
		}
	} else {
		logging.Default().Warn("sentry is not enabled")
	}

	return nil
}

func (x *Sentry) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("dsn", x.dsn),
		slog.String("env", x.env),
	)
}

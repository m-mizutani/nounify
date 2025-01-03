package cli

import (
	"context"
	"os"

	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/errutil"
	"github.com/m-mizutani/nounify/pkg/utils/logging"
	"github.com/urfave/cli/v2"
)

func Run(argv []string) error {
	var (
		logLevel  string
		logFormat string
	)

	app := cli.App{
		Name:    "nounify",
		Usage:   "Universal Slack notification tool for ALL HTTP webhooks",
		Version: types.AppVersion,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Log level (debug, info, warn, error)",
				EnvVars:     []string{"NOUNIFY_LOG_LEVEL"},
				Destination: &logLevel,
				Value:       "info",
			},
			&cli.StringFlag{
				Name:        "log-format",
				Usage:       "Log format (console, json)",
				EnvVars:     []string{"NOUNIFY_LOG_FORMAT"},
				Destination: &logFormat,
				Value:       "console",
			},
		},

		Before: func(c *cli.Context) error {
			if err := logging.Configure(os.Stdout, logLevel, logFormat); err != nil {
				return err
			}
			return nil
		},

		Commands: []*cli.Command{
			cmdServe(),
		},
	}

	if err := app.Run(argv); err != nil {
		errutil.Handle(context.Background(), "exit with failure", err)
		return err
	}

	return nil
}

package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/m-mizutani/clog"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/masq"
)

var defaultLogger *slog.Logger

func init() {
	if err := Configure(os.Stdout, "info", "console"); err != nil {
		panic(err)
	}
}

func Default() *slog.Logger {
	return defaultLogger
}

func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

func Configure(w io.Writer, level, format string) error {
	logger, err := New(w, level, format)
	if err != nil {
		return err
	}

	defaultLogger = logger
	return nil
}

func New(w io.Writer, level, format string) (*slog.Logger, error) {
	replacer := masq.New(
		masq.WithTag("secret"),
		masq.WithFieldName("Authorization"),
	)

	logLevels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	slogLevel, ok := logLevels[level]
	if !ok {
		return nil, goerr.New("invalid log level, must be debug, info, warn or error").With("level", level)
	}

	var handler slog.Handler
	switch format {
	case "console":
		handler = clog.New(
			clog.WithWriter(w),
			clog.WithLevel(slogLevel),
			clog.WithAttrHook(clog.GoerrHook),
			clog.WithReplaceAttr(replacer),
			clog.WithSource(true),
			clog.WithColorMap(&clog.ColorMap{
				Level: map[slog.Level]*color.Color{
					slog.LevelDebug: color.New(color.FgGreen, color.Bold),
					slog.LevelInfo:  color.New(color.FgCyan, color.Bold),
					slog.LevelWarn:  color.New(color.FgYellow, color.Bold),
					slog.LevelError: color.New(color.FgRed, color.Bold),
				},
				LevelDefault: color.New(color.FgBlue, color.Bold),
				Time:         color.New(color.FgWhite),
				Message:      color.New(color.FgHiWhite),
				AttrKey:      color.New(color.FgHiCyan),
				AttrValue:    color.New(color.FgHiWhite),
			}),
		)

	case "json":
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slogLevel,
			ReplaceAttr: replacer,
		})

	default:
		return nil, goerr.New("Unknown log format, must be either one of 'console' or 'json'").With("format", format)
	}

	return slog.New(handler), nil
}

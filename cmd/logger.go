package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/torgiren/ksef-cli/pkg/ksef"
)

func setupLogger(verbosity int) {
	var logLevel slog.Level
	switch verbosity {
	case 0:
		logLevel = ksef.LevelWarn
	case 1:
		logLevel = ksef.LevelInfo
	case 2:
		logLevel = ksef.LevelDebug
	case 3:
		logLevel = ksef.LevelTrace
	default:
		logLevel = ksef.LevelSecret
	}

	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "NIP" || a.Key == "nip" || a.Key == "currentNip" {
				return tint.Attr(14, a)
			}
			if a.Key == "endpoint" || a.Key == "api" {
				return tint.Attr(3, a)
			}
			if a.Key == slog.LevelKey && len(groups) == 0 {
				level, ok := a.Value.Any().(slog.Level)
				if ok && level <= ksef.LevelSecret {
					return tint.Attr(9, slog.String(a.Key, "SEC"))
				}
				if ok && level <= ksef.LevelTrace {
					return tint.Attr(12, slog.String(a.Key, "TRC"))
				}
				if ok && level <= ksef.LevelDebug {
					return tint.Attr(4, slog.String(a.Key, "DBG"))
				}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))
}

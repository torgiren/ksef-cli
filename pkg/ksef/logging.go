package ksef

import "log/slog"

const (
	LevelWarn   = slog.LevelWarn
	LevelInfo   = slog.LevelInfo
	LevelDebug  = slog.LevelDebug
	LevelTrace  = slog.LevelDebug - 4
	LevelSecret = slog.LevelDebug - 8
)

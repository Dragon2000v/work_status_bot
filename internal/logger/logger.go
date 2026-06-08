package logger

import (
	"io"
	"log/slog"
	"strings"
)

const (
	EnvLocal      = "local"
	EnvProduction = "production"
)

func New(out io.Writer, appEnv, logLevel string) (*slog.Logger, error) {
	level, err := ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: level}
	if appEnv == EnvProduction {
		return slog.New(slog.NewJSONHandler(out, opts)), nil
	}
	return slog.New(slog.NewTextHandler(out, opts)), nil
}

func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "", "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, ErrInvalidLogLevel
	}
}

var ErrInvalidLogLevel = errString("invalid log level")

type errString string

func (e errString) Error() string { return string(e) }

func RedactSecrets(input string, secrets ...string) string {
	out := input
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[REDACTED]")
	}
	return out
}

func SafeError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

package logger

import (
	"io"
	"log/slog"
	"os"
)

type LoggerBuilder struct {
	local     bool
	addSource bool
	lvl       slog.Level
	writers   []io.Writer
}

func NewBuilder() *LoggerBuilder {
	return new(LoggerBuilder)
}

func (b *LoggerBuilder) WithWriter(w io.Writer) *LoggerBuilder {
	b.writers = append(b.writers, w)

	return b
}

func (b *LoggerBuilder) WithLevel(l slog.Level) *LoggerBuilder {
	b.lvl = l

	return b
}

func (b *LoggerBuilder) Local() *LoggerBuilder {
	b.local = true

	return b
}

func (b *LoggerBuilder) WithSource() *LoggerBuilder {
	b.addSource = true

	return b
}

func (b *LoggerBuilder) Build() *slog.Logger {
	if len(b.writers) == 0 {
		b.writers = append(b.writers, os.Stdout)
	}

	w := io.MultiWriter(b.writers...)

	if b.local {
		opts := PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level:     b.lvl,
				AddSource: b.addSource,
			},
		}

		handler := opts.NewPrettyHandler(w)

		return slog.New(handler)
	}

	return newLogger(b.lvl, w)
}

func newLogger(lvl slog.Level, w io.Writer) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl}),
	)
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func MapLevel(lvl string) slog.Level {
	switch lvl {
	case "dev", "local", "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

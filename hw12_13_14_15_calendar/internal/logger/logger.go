package logger

import (
	"io"
	"log/slog"
)

// New создаёт новый логгер с заданным уровнем логирования и параметром serviceName.
// Возвращает сам логгер и levelVar для динамического изменения уровня логирования логгера.
func New(w io.Writer, level slog.Level, serviceName string) (*slog.Logger, *slog.LevelVar) {
	levelVar := &slog.LevelVar{}
	levelVar.Set(level)

	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   false,
		Level:       levelVar,
		ReplaceAttr: nil,
	})

	logger := slog.New(h)
	logger = logger.With(slog.String("service", serviceName))

	return logger, levelVar
}

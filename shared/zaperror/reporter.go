package zaperror

import (
	"context"

	"go.uber.org/zap"
)

type Reporter struct {
	logger *zap.Logger
}

func NewReporter(logger *zap.Logger) *Reporter {
	return &Reporter{logger: logger}
}

func (r *Reporter) Report(_ context.Context, err error, attrs map[string]string) {
	fields := make([]zap.Field, 0, len(attrs)+1)
	for k, v := range attrs {
		fields = append(fields, zap.String(k, v))
	}
	fields = append(fields, zap.Error(err))
	r.logger.Warn("reported error", fields...)
}

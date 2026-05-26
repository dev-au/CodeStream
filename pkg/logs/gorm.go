package logs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	slowThreshold = 200 * time.Millisecond
)

type gormLogger struct {
	zapLogger     *zap.Logger
	slowThreshold time.Duration
}

func NewGormLogger() gormlogger.Interface {
	return &gormLogger{
		zapLogger:     packageLogger.WithOptions(zap.AddCallerSkip(3)),
		slowThreshold: slowThreshold,
	}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...any) {
	FromContext(ctx).Info(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...any) {
	FromContext(ctx).Warn(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...any) {
	FromContext(ctx).Error(fmt.Sprintf(msg, data...))
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	log := FromContext(ctx).WithOptions(zap.AddCallerSkip(2))

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fields = append(fields, zap.Error(err))
		log.Error("GORM Trace Error", fields...)
		return
	}

	if elapsed > l.slowThreshold {
		log.Warn(fmt.Sprintf("SLOW SQL >= %v", l.slowThreshold), fields...)
		return
	}

	log.Debug(sql,
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	)
}

package logs

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dev-au/CodeStream/pkg/customerrors"
)

var packageLogger *zap.Logger

func Init(mode, logLevel string) {
	var zapConfig zap.Config

	if mode == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}
	var err error

	zapConfig.Level, err = zap.ParseAtomicLevel(logLevel)
	if err != nil {
		panic(err)
	}

	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	packageLogger, err = zapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(msg string, fields ...any) {
	packageLogger.Info(msg, toZapFields(fields)...)
}

func Warn(msg string, fields ...any) {
	packageLogger.Warn(msg, toZapFields(fields)...)
}

func Error(msg string, fields ...any) {
	packageLogger.Error(msg, toZapFields(fields)...)
}

func Fatal(msg string, fields ...any) {
	packageLogger.Fatal(msg, toZapFields(fields)...)
}

func Panic(msg string, fields ...any) {
	packageLogger.Panic(msg, toZapFields(fields)...)
}

func Debug(msg string, fields ...any) {
	packageLogger.Debug(msg, toZapFields(fields)...)
}

func Sync() {
	if packageLogger != nil {
		_ = packageLogger.Sync()
	}
}

func With(fields ...any) *zap.Logger {
	return packageLogger.With(toZapFields(fields)...)
}

type contextKey string

const requestIDKey contextKey = "request_id"

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return packageLogger
	}

	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok || requestID == "" {
		return packageLogger
	}

	return packageLogger.With(zap.String("request_id", requestID))
}

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok || requestID == "" {
		return ""
	}
	return requestID
}

func InfoCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Info(msg, toZapFields(fields)...)
}

func ErrorCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Error(msg, toZapFields(fields)...)
}

func WarnCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Warn(msg, toZapFields(fields)...)
}

func DebugCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Debug(msg, toZapFields(fields)...)
}

func PanicCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Panic(msg, toZapFields(fields)...)
}

func FatalCtx(ctx context.Context, msg string, fields ...any) {
	FromContext(ctx).Fatal(msg, toZapFields(fields)...)
}

func LogCtx(ctx context.Context, err error, msg string, fields ...any) {
	fields = append(fields, "error", err)
	if errors.Is(err, customerrors.ErrExternalService) || errors.Is(err, customerrors.ErrDataNotFound) {
		FromContext(ctx).Warn(msg, toZapFields(fields)...)
	} else {
		FromContext(ctx).Error(msg, toZapFields(fields)...)
	}
}

func toZapFields(fields []any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			zapFields = append(zapFields, zap.Any(key, fields[i+1]))
		}
	}
	return zapFields
}

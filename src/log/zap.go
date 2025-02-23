package log

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Zap struct {
	logger *zap.Logger
}

func NewZap() *Zap {
	return &Zap{
		logger: newZap(),
	}
}

func newZap() *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:  "_m",
		NameKey:     "logger",
		LevelKey:    "_l",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		TimeKey:     "_t",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	}

	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zap.DebugLevel))
}

func (z *Zap) Info(msg string, args ...zap.Field) {
	z.logger.Info(msg, args...)
}

func (z *Zap) Warn(msg string, args ...zap.Field) {
	z.logger.Warn(msg, args...)
}

func (z *Zap) Error(msg string, args ...zap.Field) {
	z.logger.Error(msg, args...)
}

func (z *Zap) With(args ...string) *Zap {
	zapArgs := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		zapArgs = append(zapArgs, zap.String("", arg))
	}

	return &Zap{
		logger: z.logger.With(zapArgs...),
	}
}

func (z *Zap) Close() error {
	if err := z.logger.Sync(); err != nil && !isSyncInvalidError(err) {
		return fmt.Errorf("failed sync logger | %w", err)
	}

	return nil
}

func isSyncInvalidError(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) && (errors.Is(pathErr.Err, syscall.ENOTTY) || errors.Is(pathErr.Err, syscall.EINVAL)) {
		return true
	}

	return false
}

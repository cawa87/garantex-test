package sl

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(level string) (*Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: logger}, nil
}

func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Errorw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Warnw(msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Infow(msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Debugw(msg, keysAndValues...)
}

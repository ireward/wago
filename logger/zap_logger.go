package logger

import (
	"fmt"
	"os"
	"strings"

	"wago/logger/tag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	zl   *zap.Logger
	skip int
}

var _ Logger = (*zapLogger)(nil)

// NewDefaultLogger returns a logger at debug level and log
// into StdErr.
func NewDefaultLogger() *zapLogger {
	return NewZapLogger(BuildZapLogger(Config{
		Level: "debug",
	}))
}

// NewLoggerFromConf builds and returns a new zap based logger
// from zap.Logger for this logging configuration.
func NewZapLoggerFromConf(conf Config) *zapLogger {
	return NewZapLogger(BuildZapLogger(conf))
}

// NewCLI Logger builds and returns a new zap based logger from
// zap.Logger used for cli-logging.
func NewCLIZapLogger() *zapLogger {
	return NewZapLogger(buildCLIZapLogger())
}

// NewZapLogger returns a new zap based logger from zap.Logger
func NewZapLogger(zl *zap.Logger) *zapLogger {
	return &zapLogger{zl: zl, skip: 3}
}

// BuildZapLogger builds and returns a new zap.Logger for this logging
// configuration.
func BuildZapLogger(cfg Config) *zap.Logger {
	return buildZapLogger(cfg, true)
}

func (zl *zapLogger) buildFieldsWithCallAt(tags []tag.Tag) []zap.Field {
	fields := make([]zap.Field, len(tags)+1)
	zl.fillFields(tags, fields)
	fields[len(fields)-1] = zap.String(loggingCallAtKey, caller(zl.skip))
	return fields
}

func (l *zapLogger) fillFields(tags []tag.Tag, fields []zap.Field) {
	for i, t := range tags {
		if zt, ok := t.(tag.ZapTag); ok {
			fields[i] = zt.Field()
		} else {
			fields[i] = zap.Any(t.Key(), t.Value())
		}
	}
}

func (l *zapLogger) Debug(msg string, tags ...tag.Tag) {
	if l.zl.Core().Enabled(zap.DebugLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFieldsWithCallAt(tags)
		l.zl.Debug(msg, fields...)
	}
}

func (l *zapLogger) Info(msg string, tags ...tag.Tag) {
	if l.zl.Core().Enabled(zap.InfoLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFieldsWithCallAt(tags)
		l.zl.Info(msg, fields...)
	}
}

func (l *zapLogger) Warn(msg string, tags ...tag.Tag) {
	if l.zl.Core().Enabled(zap.WarnLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFieldsWithCallAt(tags)
		l.zl.Warn(msg, fields...)
	}
}

func (l *zapLogger) Error(msg string, tags ...tag.Tag) {
	if l.zl.Core().Enabled(zap.ErrorLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFieldsWithCallAt(tags)
		l.zl.Error(msg, fields...)
	}
}

func (l *zapLogger) Fatal(msg string, tags ...tag.Tag) {
	if l.zl.Core().Enabled(zap.FatalLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFieldsWithCallAt(tags)
		l.zl.Fatal(msg, fields...)
	}
}

func (l *zapLogger) With(tags ...tag.Tag) Logger {
	fields := make([]zap.Field, len(tags))
	l.fillFields(tags, fields)
	zl := l.zl.With(fields...)
	return &zapLogger{
		zl:   zl,
		skip: l.skip,
	}
}

func (l *zapLogger) Skip(extraSkip int) Logger {
	return &zapLogger{
		zl:   l.zl,
		skip: l.skip + extraSkip,
	}
}

func buildZapLogger(cfg Config, disableCaller bool) *zap.Logger {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		LevelKey:       levelKey,
		NameKey:        nameKey,
		CallerKey:      zapcore.OmitKey, // we use our own caller
		MessageKey:     messageKey,
		StacktraceKey:  stackTraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if disableCaller {
		encodeConfig.CallerKey = zapcore.OmitKey
		encodeConfig.EncodeCaller = nil
	}

	outputPath := "stderr"
	if len(cfg.OutputFile) > 0 {
		file, err := os.Open(cfg.OutputFile)
		if err != nil {
			panic(err)
		}
		fileInfo, err := file.Stat()
		if err != nil {
			panic(err)
		}

		if fileInfo.IsDir() {
			cfg.OutputFile = fmt.Sprintf("%s/%s", cfg.OutputFile, "dcm_app.log")
		}
		outputPath = cfg.OutputFile
	}

	if cfg.Stdout {
		outputPath = "stdout"
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(parseZapLevel(cfg.Level)),
		Development:      false,
		Sampling:         nil,
		Encoding:         "json",
		EncoderConfig:    encodeConfig,
		OutputPaths:      []string{outputPath},
		ErrorOutputPaths: []string{outputPath},
		DisableCaller:    disableCaller,
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func buildCLIZapLogger() *zap.Logger {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		LevelKey:       levelKey,
		NameKey:        nameKey,
		CallerKey:      zapcore.OmitKey,
		MessageKey:     messageKey,
		StacktraceKey:  stackTraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   nil,
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableStacktrace: os.Getenv("dcm_CLI_SHOW_STACKS") == "",
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encodeConfig,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
	}
	logger, _ := config.Build()
	return logger

}

func parseZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

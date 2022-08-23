package logger

import (
	"wago/logger/tag"
)

type withLogger struct {
	logger Logger
	tags   []tag.Tag
}

var _ Logger = (*withLogger)(nil)

// With returns Logger instance that prepend every log entry with tags. If logger implements WithLogger it is used,
// otherwise every log call will be intercepted
func With(logger Logger, tags ...tag.Tag) Logger {
	if w, ok := logger.(WithLogger); ok {
		return w.With(tags...)
	}
	return newWithLogger(logger, tags...)
}

func newWithLogger(logger Logger, tags ...tag.Tag) *withLogger {
	return &withLogger{logger: logger, tags: tags}
}

func (l *withLogger) prependTags(tags []tag.Tag) []tag.Tag {
	return append(l.tags, tags...)
}

// Debug writes message to the log.
func (l *withLogger) Debug(msg string, tags ...tag.Tag) {
	l.logger.Debug(msg, l.prependTags(tags)...)
}

// Info writes message to the log.
func (l *withLogger) Info(msg string, tags ...tag.Tag) {
	l.logger.Info(msg, l.prependTags(tags)...)
}

// Warn writes message to the log.
func (l *withLogger) Warn(msg string, tags ...tag.Tag) {
	l.logger.Warn(msg, l.prependTags(tags)...)
}

// Error writes message to the log.
func (l *withLogger) Error(msg string, tags ...tag.Tag) {
	l.logger.Error(msg, l.prependTags(tags)...)
}

// Fatal writes message to the log.
func (l *withLogger) Fatal(msg string, tags ...tag.Tag) {
	l.logger.Fatal(msg, l.prependTags(tags)...)
}

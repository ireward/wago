package log

import "wago/log/tag"

type (
	noopLogger struct{}
)

func NewNoopLogger() *noopLogger {
	return &noopLogger{}
}

func (n *noopLogger) Debug(string, ...tag.Tag) {}
func (n *noopLogger) Info(string, ...tag.Tag)  {}
func (n *noopLogger) Warn(string, ...tag.Tag)  {}
func (n *noopLogger) Error(string, ...tag.Tag) {}
func (n *noopLogger) Fatal(string, ...tag.Tag) {}
func (n *noopLogger) With(...tag.Tag) Logger {
	return n
}

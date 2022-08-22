package log

import (
	"wago/log/tag"
)

// Logger is the logging interface.
// Note: msg should not be static, do not use fmt.Sprintf() for msg. Anything dynamic should be tagged
type (
	Logger interface {
		Debug(msg string, tags ...tag.Tag)
		Info(msg string, tags ...tag.Tag)
		Warn(msg string, tags ...tag.Tag)
		Error(msg string, tags ...tag.Tag)
		Fatal(msg string, tags ...tag.Tag)
	}
	// Implement WithLogger interface with With method should return new instance of logger with prepended tags.
	WithLogger interface {
		With(tags ...tag.Tag) Logger
	}
	// If logger implements SkipLogger then Skip method will be called and skip parameter
	// will have number of extra stack trace frames to skip (useful to log calle func file/line)
	SkipLogger interface {
		Skip(skip int) Logger
	}
)

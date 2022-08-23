package logger

// Config contains the config items for logger
type Config struct {
	// Stdout is true if the output needs to go to standard out
	Stdout bool
	// Level is the desired log level
	Level string
	// OutputFile is the path to the log output file
	OutputFile string
}

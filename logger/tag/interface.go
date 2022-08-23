package tag

type (
	// Tag is an interface to supply custom tags to Logger interface implementation.
	Tag interface {
		Key() string
		Value() interface{}
	}
)

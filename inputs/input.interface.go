package inputs

type ValueInputInterface interface {
	GetBoolean(key string) (bool, error)
	GetNumber(key string) (float64, error)
	GetString(key string) (string, error)
	Has(key string) bool

	// CanRefresh whether the implementation can
	// implement auto-refreshing at runtime
	// e.g. ENV Input cannot handle auto-refreshing
	// at runtime
	// @todo not working properly
	CanRefresh() bool

	Reload() error

	// GetInputName it simply returns current input source name
	GetInputName() string
}

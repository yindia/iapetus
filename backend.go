package iapetus

// Backend is the interface for task execution plugins.
type Backend interface {
	RunTask(task *Task) error
	ValidateTask(task *Task) error
}

// backendRegistry holds all registered backends by name.
var backendRegistry = map[string]Backend{}

// RegisterBackend registers a backend plugin by name.
// Plugin authors: call this in your plugin's init() function.
func RegisterBackend(name string, backend Backend) {
	backendRegistry[name] = backend
}

// GetBackend retrieves a backend by name, or nil if not found.
func GetBackend(name string) Backend {
	if b, ok := backendRegistry[name]; ok {
		return b
	}
	return nil
}

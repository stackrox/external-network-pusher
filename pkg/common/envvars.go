package common

var (
	verbose = false
)

// Verbose returns if verbose options is set
func Verbose() bool {
	return verbose
}

// SetVerbose sets verbose option
func SetVerbose() {
	verbose = true
}

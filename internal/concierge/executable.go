package concierge

import "fmt"

const (
	RestoreAction string = "restore"
	PrepareAction string = "prepare"
)

// Executable is an interface that represents any struct implementing the Prepare/Restore methods.
type Executable interface {
	Prepare() error
	Restore() error
}

// DoAction takes an Executable, and calls either Prepare() or Restore() according
// to the action parameter.
func DoAction(executable Executable, action string) error {
	switch action {
	case PrepareAction:
		return executable.Prepare()
	case RestoreAction:
		return executable.Restore()
	default:
		return fmt.Errorf("unknown executor action: %s", action)
	}
}

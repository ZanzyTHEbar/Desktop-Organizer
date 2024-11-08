package terminal

// Terminal interface defines methods for terminal operations
type TerminalI interface {
	OutputErrorAndExit(msg string, args ...interface{})
	OutputInfo(msg string, args ...interface{})
	OutputSuccess(msg string, args ...interface{})
	ConfirmYesNo(msg string) bool
	ToggleSpinner()
}

package terminal

// Terminal interface defines methods for terminal operations
type Terminal interface {
	OutputErrorAndExit(format string, a ...interface{})
	OutputInfo(format string, a ...interface{})
	OutputSuccess(format string, a ...interface{})
}

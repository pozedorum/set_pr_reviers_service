package interfaces

type Logger interface {
	Debug(operation, message string, keyvals ...interface{})
	Info(operation, message string, keyvals ...interface{})
	Warn(operation, message string, keyvals ...interface{})
	Error(operation, message string, keyvals ...interface{})
	Shutdown()
}

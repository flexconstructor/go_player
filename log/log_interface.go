package log
// Interface represents logger with different levels of logs.
type Logger interface {
	//Debug layer.
	Debug(interface{}, ...interface{})
	//Trace layer
	Trace(interface{}, ...interface{})
	//Info layer
	Info(interface{}, ...interface{})
	//Warning layer
	Warn(interface{}, ...interface{}) error
	//Error layer
	Error(interface{}, ...interface{}) error
	//Critical layer
	Critical(interface{}, ...interface{}) error
	//Closing method
	Close()
}

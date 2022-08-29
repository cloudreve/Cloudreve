package logger

var GlobalLogger = NewLogger()

var (
	Debug   = GlobalLogger.Debugf
	Info    = GlobalLogger.Infof
	Warning = GlobalLogger.Warnf
	Error   = GlobalLogger.Errorf
	Panic   = GlobalLogger.Panicf
	Println = GlobalLogger.Println
)

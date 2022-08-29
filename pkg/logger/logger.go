package logger

var GlobalLogger = NewLogger()

var (
	Debug   = GlobalLogger.Debugf
	Info    = GlobalLogger.Infof
	Warn    = GlobalLogger.Warnf
	Warning = GlobalLogger.Warnf
	Error   = GlobalLogger.Errorf
	Fatal   = GlobalLogger.Fatalf
	Panic   = GlobalLogger.Panicf
	Println = GlobalLogger.Println
)

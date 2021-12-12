package taskflow

type Logger interface {
	Printf(format string, v ...interface{})
}

type NoopLogger struct {
}

func (n NoopLogger) Printf(format string, v ...interface{}) {
}

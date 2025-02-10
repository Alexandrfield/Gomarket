package common

type Logger interface {
	Warnf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

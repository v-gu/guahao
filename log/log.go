package log

import (
	glog "github.com/golang/glog"
)

const (
	DEBUG_CONFIG glog.Level = iota + 1
	DEBUG_SESSION
	DEBUG_HTTP
)

func Info(args ...interface{}) {
}

type NamedLogger struct {
	Name string
}

//
func (m *NamedLogger) Info(args ...interface{}) {
	glog.Info(m.Name+": ", args)
}
func (m *NamedLogger) Infoln(args ...interface{}) {
	glog.Infoln(m.Name+": ", args)
}
func (m *NamedLogger) Infof(format string, args ...interface{}) {
	glog.Infof("%s: "+format, m.Name, args)
}

//
func (m *NamedLogger) Warning(args ...interface{}) {
	glog.Warning(m.Name+": ", args)
}
func (m *NamedLogger) Warningln(args ...interface{}) {
	glog.Warningln(m.Name+": ", args)
}
func (m *NamedLogger) Warningf(format string, args ...interface{}) {
	glog.Warningf("%s: "+format, args)
}

//
func (m *NamedLogger) Error(args ...interface{}) {
	glog.Error(m.Name+": ", args)
}
func (m *NamedLogger) Errorln(args ...interface{}) {
	glog.Errorln(m.Name+": ", args)
}
func (m *NamedLogger) Errorf(format string, args ...interface{}) {
	glog.Errorf("%s: "+format, args)
}

//
func (m *NamedLogger) Debug(level glog.Level, args ...interface{}) {
	if glog.V(level) {
		m.Info(args)
	}
}
func (m *NamedLogger) Debugln(level glog.Level, args ...interface{}) {
	if glog.V(level) {
		m.Infoln(args)
	}
}
func (m *NamedLogger) Debugf(level glog.Level, format string, args ...interface{}) {
	if glog.V(level) {
		m.Infof(format, args)
	}
}

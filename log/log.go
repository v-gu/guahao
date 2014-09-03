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
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Info(nargs...)
}
func (m *NamedLogger) Infoln(args ...interface{}) {
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Infoln(nargs...)
}
func (m *NamedLogger) Infof(format string, args ...interface{}) {
	nargs := []interface{}{m.Name}
	nargs = append(nargs, args...)
	glog.Infof("%s: "+format, nargs...)
}

//
func (m *NamedLogger) Warning(args ...interface{}) {
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Warning(nargs...)
}
func (m *NamedLogger) Warningln(args ...interface{}) {
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Warningln(nargs...)
}
func (m *NamedLogger) Warningf(format string, args ...interface{}) {
	nargs := []interface{}{m.Name}
	nargs = append(nargs, args...)
	glog.Warningf("%s: "+format, nargs...)
}

//
func (m *NamedLogger) Error(args ...interface{}) {
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Error(nargs...)
}
func (m *NamedLogger) Errorln(args ...interface{}) {
	nargs := []interface{}{m.Name + ": "}
	nargs = append(nargs, args...)
	glog.Errorln(nargs...)
}
func (m *NamedLogger) Errorf(format string, args ...interface{}) {
	nargs := []interface{}{m.Name}
	nargs = append(nargs, args...)
	glog.Errorf("%s: "+format, nargs...)
}

//
func (m *NamedLogger) Debug(level glog.Level, args ...interface{}) {
	if glog.V(level) {
		m.Info(args...)
	}
}
func (m *NamedLogger) Debugln(level glog.Level, args ...interface{}) {
	if glog.V(level) {
		m.Infoln(args...)
	}
}
func (m *NamedLogger) Debugf(level glog.Level, format string, args ...interface{}) {
	if glog.V(level) {
		m.Infof(format, args...)
	}
}

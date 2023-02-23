package bl

import (
	"github.com/donscoco/gateway_be/pkg/iron_log"
	"path"
	"runtime"
	"strconv"
)

// 封装 logger，
// internal 中的代码就使用 这个包下的log，这个包log的具体实现再用外部的log包

func Info(arg ...interface{}) {
	iron_log.Info(arg...)
}
func Infof(format string, arg ...interface{}) {
	iron_log.Infof(format, arg...)
}
func Error(arg ...interface{}) {
	// 一些业务需要的处理

	// 添加日志打印代码位置
	location := ""
	_, file, line, ok := runtime.Caller(1)
	if ok {
		location = path.Base(file) + ":" + strconv.Itoa(line)
	}
	arg = append(arg, " code:"+location)

	iron_log.Error(arg...)
}
func Errorf(format string, arg ...interface{}) {
	// 一些处理

	// 添加日志打印代码位置
	location := ""
	_, file, line, ok := runtime.Caller(1)
	if ok {
		location = path.Base(file) + ":" + strconv.Itoa(line)
	}

	iron_log.Errorf(format+" code:"+location, arg...)
}
func Warn(arg ...interface{}) {
	iron_log.Warn(arg...)
}
func Warnf(format string, arg ...interface{}) {
	iron_log.Warnf(format, arg...)
}
func Debug(arg ...interface{}) {
	// 一些处理

	// 添加日志打印代码位置
	location := ""
	_, file, line, ok := runtime.Caller(1)
	if ok {
		location = path.Base(file) + ":" + strconv.Itoa(line)
	}
	arg = append(arg, " code:"+location)

	iron_log.Debug(arg...)
}
func Debugf(format string, arg ...interface{}) {

	// 一些处理

	// 添加日志打印代码位置
	location := ""
	_, file, line, ok := runtime.Caller(1)
	if ok {
		location = path.Base(file) + ":" + strconv.Itoa(line)
	}

	iron_log.Debugf(format+" code:"+location, arg...)
}

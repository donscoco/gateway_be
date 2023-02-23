package iron_core

import (
	"log"
	"syscall"
	"testing"
	"time"
)

// go test -v -run TestCore ./core
func TestCore(t *testing.T) {

	GoCore := NewCore()
	GoCore.SetLogger(new(TestLogger))

	go func() { // ci 下可能会卡住，2秒后自动模拟退出信号
		time.Sleep(2 * time.Second)
		GoCore.signal <- syscall.SIGINT
	}()

	GoCore.OnStart(startFunc1, startFunc2)
	GoCore.OnStop(stopFunc2, stopFunc1)
	GoCore.Boot()
}
func startFunc1() error {
	log.Println("init server1")
	return nil
}
func stopFunc1() error {
	log.Println("close server1")
	return nil
}

func startFunc2() error {
	log.Println("init server2")
	return nil
}
func stopFunc2() error {
	log.Println("close server2")
	return nil
}

type TestLogger struct{}

func (t TestLogger) Info(arg ...interface{})                  { log.Println(arg...) }
func (t TestLogger) Infof(format string, arg ...interface{})  { log.Printf(format, arg...) }
func (t TestLogger) Debug(arg ...interface{})                 {}
func (t TestLogger) Debugf(format string, arg ...interface{}) {}
func (t TestLogger) Warn(arg ...interface{})                  {}
func (t TestLogger) Warnf(format string, arg ...interface{})  {}
func (t TestLogger) Error(arg ...interface{})                 {}
func (t TestLogger) Errorf(format string, arg ...interface{}) {}

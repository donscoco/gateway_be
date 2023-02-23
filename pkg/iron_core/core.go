package iron_core

import (
	"os"
	"os/signal"
	"syscall"
)

type Logger interface {
	Info(arg ...interface{})
	Infof(format string, arg ...interface{})
	Debug(arg ...interface{})
	Debugf(format string, arg ...interface{})
	Warn(arg ...interface{})
	Warnf(format string, arg ...interface{})
	Error(arg ...interface{})
	Errorf(format string, arg ...interface{})
}

type Core struct {
	logger Logger
	signal chan os.Signal

	startActs []func() error
	stopActs  []func() error
}

func NewCore() (core *Core) {
	core = new(Core)
	core.signal = make(chan os.Signal, 1)
	return
}

func (a *Core) SetLogger(logger Logger) {
	a.logger = logger
}

// Boot 运行服务器消息监听循环，等待外部进程消息
func (a *Core) Boot(arg ...func()) {
	// 执行开机操作
	a.start()

	// 等待系统信号
	a.loop()

	// 优雅退出
	a.stop()

	return
}

// 设置开机动作
func (a *Core) OnStart(starts ...func() error) {
	a.startActs = starts
}

// 设置关机动作
func (a *Core) OnStop(stops ...func() error) {
	a.stopActs = stops
}

func (c *Core) start() {

	c.logger.Infof("应用开始启动")

	//todo 启动core的组件

	// 执行业务开机动作
	for _, startAct := range c.startActs {
		err := startAct()
		if err != nil {
			c.logger.Errorf("启动出错：%s", err)
			os.Exit(1)
		}
	}

	c.logger.Infof(`应用启动完成[PID=%d]`, os.Getpid())
}
func (c *Core) loop() {

	signal.Notify(c.signal, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM) // 监听信号

	for {
		sig := <-c.signal
		c.logger.Infof("接收到信号: %s", sig)

		switch sig {
		case syscall.SIGUSR1: // 自定义操作，例如设置为debug日志模式
			// todo
		case syscall.SIGUSR2:
			// todo
		default: // sigint 和 sigterm 就退出
			return
		}
	}

}
func (c *Core) stop() {
	c.logger.Infof("应用开始关闭")

	var err error
	// 执行业务关闭动作
	for _, stopAct := range c.stopActs {
		err = stopAct()
		if err != nil {
			c.logger.Errorf("关闭出错: %s", err)
		}
	}

	//todo 关闭core的组件
	//mlog.Close()

	c.logger.Infof("应用关闭完成")
}

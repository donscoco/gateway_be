package mkafka

import (
	"fmt"
	"github.com/donscoco/gochat/pkg/iron_log"
)

type saramaLogger struct {
	logger *iron_log.ServerLogger
}

func (s *saramaLogger) Print(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}
func (s *saramaLogger) Printf(format string, v ...interface{}) {
	s.logger.Infof(format, v...)
}
func (s *saramaLogger) Println(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}

// 可以在应用打印 kafka的相关信息
func newSaramaLogger() (sl *saramaLogger) {
	sl = &saramaLogger{
		logger: iron_log.NewLogger("saramaLogger"),
	}

	return
}

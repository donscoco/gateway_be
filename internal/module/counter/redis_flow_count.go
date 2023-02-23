package counter

import (
	"errors"
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/pkg/base/redis"

	//"github.com/garyburd/redigo/redis"
	"sync/atomic"
	"time"
)

type RedisFlowCountService struct {
	AppID       string
	Interval    time.Duration
	QPS         int64
	Unix        int64
	TickerCount int64
	TotalCount  int64
}

func NewRedisFlowCountService(appID string, interval time.Duration, loc *time.Location) *RedisFlowCountService {
	reqCounter := &RedisFlowCountService{
		AppID:    appID,
		Interval: interval,
		QPS:      0,
		Unix:     0,
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount) //获取数据
			atomic.StoreInt64(&reqCounter.TickerCount, 0)            //重置数据

			currentTime := time.Now()
			dayKey := reqCounter.GetDayKey(currentTime, loc)
			hourKey := reqCounter.GetHourKey(currentTime, loc)

			p, ok := redis.RedisSingleClients["default"]
			if !ok {
				//todo log
				bl.Error("[counter] get redis fail")
				return
			}
			pipe := p.Pipeline()

			pipe.IncrBy(dayKey, tickerCount)
			pipe.Expire(dayKey, time.Hour*24*2)
			pipe.IncrBy(hourKey, tickerCount)
			pipe.Expire(hourKey, time.Hour*24*2)

			_, err := pipe.Exec()
			if err != nil {
				fmt.Println("RedisConfPipline err", err)
				continue
			}

			totalCount, err := reqCounter.GetDayData(currentTime, loc)
			if err != nil {
				fmt.Println("reqCounter.GetDayData err", err)
				continue
			}
			nowUnix := time.Now().Unix()
			if reqCounter.Unix == 0 {
				reqCounter.Unix = time.Now().Unix()
				continue
			}
			tickerCount = totalCount - reqCounter.TotalCount
			if nowUnix > reqCounter.Unix {
				reqCounter.TotalCount = totalCount
				reqCounter.QPS = tickerCount / (nowUnix - reqCounter.Unix)
				reqCounter.Unix = time.Now().Unix()
			}
		}
	}()
	return reqCounter
}

func (o *RedisFlowCountService) GetDayKey(t time.Time, loc *time.Location) string {
	dayStr := t.In(loc).Format("20060102")
	return fmt.Sprintf("%s_%s_%s", bl.RedisFlowDayKey, dayStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourKey(t time.Time, loc *time.Location) string {
	hourStr := t.In(loc).Format("2006010215")
	return fmt.Sprintf("%s_%s_%s", bl.RedisFlowHourKey, hourStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourData(t time.Time, loc *time.Location) (int64, error) {

	hourKey := o.GetHourKey(t, loc)

	p, ok := redis.RedisSingleClients["default"]
	if !ok {
		//todo log
		bl.Error("[counter] get redis fail")
		return 0, errors.New("get redis fail")
	}

	return p.Get(hourKey).Int64()

}

func (o *RedisFlowCountService) GetDayData(t time.Time, loc *time.Location) (int64, error) {

	dayKey := o.GetDayKey(t, loc)

	p, ok := redis.RedisSingleClients["default"]
	if !ok {
		//todo log
		bl.Error("[counter] get redis fail")
		return 0, errors.New("get redis fail")
	}

	return p.Get(dayKey).Int64()
}

// 原子增加
func (o *RedisFlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&o.TickerCount, 1)
	}()
}

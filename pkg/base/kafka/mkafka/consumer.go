package mkafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"os"
	"strconv"
	"sync"
	"time"
)

type KafkaConsumer struct {
	config struct {
		Name    string
		Brokers []string // 服务器列表
		SASL    struct {
			Enable   bool // 是否启用加密信息，没有密码的服务器不要设置为true，即使不填user和pass也连不上
			User     string
			Password string
		}
		ClientId   string
		GroupId    string   // 消费组
		Topics     []string // topic
		AutoCommit struct {
			Enable   bool
			Interval int
		}
		SetLog  bool
		Version string // 驱动应该用什么版本的api与服务器对话，最好和服务器版本保持一致以避免怪问题
	}
	saramaConfig *cluster.Config
	consumer     *cluster.Consumer
	//output       chan *sarama.ConsumerMessage

	handleFunc      func(*KafkaConsumer, *sarama.ConsumerMessage) error
	successCallback func(*KafkaConsumer, *sarama.ConsumerMessage) error
	failureCallback func(*KafkaConsumer, *sarama.ConsumerMessage) error

	sync.WaitGroup
}

func CreateConsumer(Brokers []string, GroupId string, Topics []string,
	config map[string]string,
	handleFunc func(*KafkaConsumer, *sarama.ConsumerMessage) error,
	successCallback func(*KafkaConsumer, *sarama.ConsumerMessage) error,
	failureCallback func(*KafkaConsumer, *sarama.ConsumerMessage) error,
) (c *KafkaConsumer, err error) {
	c = new(KafkaConsumer)

	cconf := cluster.NewConfig()
	cconf.ClientID = config["ClientId"]
	// 对应配置
	cconf.Consumer.Return.Errors = true
	cconf.Group.Return.Notifications = true

	// 配置自动提交, fixme sarama-cluster包还是自动提交，默认用的CommitInterval，Shopify/sarama包已经改为可以配置不用自动提交了。后续需要修改
	//cconf.Consumer.Offsets.CommitInterval = time.Duration(c.config.AutoCommit.Interval) * time.Second // 旧版本
	//if c.config.AutoCommit.Enable {
	//	cconf.Consumer.Offsets.AutoCommit.Interval = time.Duration(c.config.AutoCommit.Interval) * time.Second
	//}
	if len(config["AutoCommitInterval"]) > 0 {
		sec, _ := strconv.Atoi(config["AutoCommitInterval"])
		cconf.Consumer.Offsets.CommitInterval = time.Duration(sec) * time.Second
		cconf.Consumer.Offsets.AutoCommit.Interval = time.Duration(sec) * time.Second
		cconf.Consumer.Offsets.AutoCommit.Enable = false
	}

	if config["Version"] != "" {
		kafkaVersion, err := sarama.ParseKafkaVersion(config["Version"])
		if err == nil {
			cconf.Version = kafkaVersion
			logger.Debugf("Kafka 版本配置为: %s", kafkaVersion)
		} else {
			logger.Errorf("错误的 Kafka 版本配置: %s", config["Version"])
			os.Exit(1)
		}
	}

	c.saramaConfig = cconf
	c.consumer, err = cluster.NewConsumer(Brokers, GroupId, Topics, c.saramaConfig)
	if err != nil {
		return nil, err
	}
	//c.output = make(chan *sarama.ConsumerMessage, 100)

	if config["SetLog"] == "true" {
		sarama.Logger = newSaramaLogger()
	}

	c.handleFunc = handleFunc
	c.failureCallback = failureCallback
	c.successCallback = successCallback

	return c, nil
}
func (c *KafkaConsumer) Start() {

	go func() {
		c.Add(1)
		c.Worker()
		c.Done()
	}()
}
func (c *KafkaConsumer) Worker() {

	for {
		select {
		case msg, ok := <-c.consumer.Messages():
			if !ok {
				// 收到关闭通知,关闭下游
				//close(c.output)
				return
			}

			err := c.handleFunc(c, msg)
			if err != nil {
				c.failureCallback(c, msg)
			} else {
				c.successCallback(c, msg)
			}

			//// fixme : 做成执行完业务才能提交偏移量。但是看依赖包的mloop是自动提交的。
			//c.consumer.MarkOffset(msg, "")
			//c.consumer.CommitOffsets()
		case n, ok := <-c.consumer.Notifications():
			if ok {
				logger.Infof("%s发出通知: %s", c, n.Type)
				logger.Infof("减持分区 %+v", n.Released)
				logger.Infof("增持分区 %+v", n.Claimed)
				logger.Infof("当前持有 %+v", n.Current)
			}
		case err, ok := <-c.consumer.Errors():
			if !ok {
				return
			}
			logger.Errorf("%s出错: %s", c, err)
		}
	}

}
func (c *KafkaConsumer) Stop() {
	c.consumer.Close()

	c.Wait()
}

func (c *KafkaConsumer) String() string {
	return fmt.Sprintf("消费者(name=%s, groupId=%s, topics=%v)", c.config.Name, c.config.GroupId, c.config.Topics)
}

//func (c *KafkaConsumer) Output() chan *sarama.ConsumerMessage {
//	return c.output
//}

func (s KafkaConsumer) Consumer() (consumer *cluster.Consumer) {
	return s.consumer
}

// kafka Client 会定时提交offset，这里提供主动提交接口
func (s *KafkaConsumer) Commit(msgs ...sarama.ConsumerMessage) {
	for _, m := range msgs {
		s.consumer.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "")
	}
	s.consumer.CommitOffsets()
}

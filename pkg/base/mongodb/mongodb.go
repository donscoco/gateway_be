package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var DefaultMgoDB *mongo.Client

//var DefaultAddr = []string{"192.168.2.132:27017"} // todo 文件配置

func InitMgoClient(addr []string, timeout int) (err error) {
	cli, err := CreateClient(addr, timeout)
	if err != nil {
		return err
	}
	DefaultMgoDB = cli
	return nil
}

func CreateClient(addrs []string, timeout int) (client *mongo.Client, err error) {

	clientOpts := options.Client().
		SetConnectTimeout(time.Duration(timeout) * time.Second).
		SetHosts(
			//[]string{
			//"192.168.2.132:27017",
			////"192.168.2.132:27018",
			////"192.168.2.132:27019",
			//}
			addrs,
		)
	//SetMaxPoolSize(uint64(MaxPoolSize)).
	//SetMinPoolSize(uint64(MinPoolSize)).
	//SetReplicaSet("ironhead")
	client, err = mongo.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}
	client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return client, nil

}

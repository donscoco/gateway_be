package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"time"
)

var DefaultOSS *oss.Client

// todo ！！！ 第一版先直接写，后续改成配置
var DefaultEndpoint = "https://oss-cn-guangzhou.aliyuncs.com"

// var DefaultAccessKeyId = "LTAI5tHkB2m5AjwRej8WQp6J"
var DefaultAccessKeyId = ""
var DefaultAccessKeySecret = ""

var DefaultBucket = "donscoco-bucket" // 公共读

var DefaultURLPrefix = "https://donscoco-bucket.oss-cn-guangzhou.aliyuncs.com/"

// 初始化
func InitOSS(Endpoint, AccessKeyId, AccessKeySecret string) error {
	client, err := NewOSSClient(Endpoint, AccessKeyId, AccessKeySecret)
	if err != nil {
		return err
	}
	DefaultOSS = client
	return nil
}

func NewOSSClient(Endpoint, AccessKeyId, AccessKeySecret string) (c *oss.Client, err error) {
	// 配置
	option := func(c *oss.Client) {
		c.Config.RetryTimes = 10
		c.Config.Timeout = 60
		// ...
	}
	client, err := oss.New(Endpoint, AccessKeyId, AccessKeySecret, option)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// 添加  填写存储空间名称
func PutObject(client *oss.Client, bucketName string, key string, input io.Reader, t time.Time, aclIsPublic bool) (err error) {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	option := []oss.Option{
		//oss.Expires(expires),
		//oss.Meta("MyProp", "MyPropVal")
	}
	futureTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
	fmt.Println(futureTime)
	option = append(option, oss.Expires(futureTime))

	if aclIsPublic {
		option = append(option, oss.ObjectACL(oss.ACLPublicRead))
	}

	err = bucket.PutObject(key, input, option...)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return
}

// todo 添加临时资源 设置 过期时间 https://help.aliyun.com/document_detail/32152.html

// 查找
func GetObject(client *oss.Client, bucketName string, key string) (output io.Reader, err error) {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	option := []oss.Option{}

	output, err = bucket.GetObject(key, option...)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return

}

// 删除
func DeleteObject(client *oss.Client, bucketName string, key string) (err error) {
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	option := []oss.Option{}

	err = bucket.DeleteObject(key, option...)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return
}

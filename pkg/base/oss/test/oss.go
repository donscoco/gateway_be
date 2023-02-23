package main

import (
	"fmt"
	"github.com/donscoco/gochat/internal/base/oss"
	"os"

	//"os"
	"time"
)

// 默认域名 ironhead
var Endpoint = "https://oss-cn-guangzhou.aliyuncs.com"
var AccessKeyId = "LTAI5tHkB2m5AjwRej8WQp6J"
var AccessKeySecret = "gJEQlpRrMzhxwpQl5fRFPPk1dX2nmG"

// 登陆密码 DKU1bo1xzZyXGECCb5p2#unj(C!eW@Ks

var Bucket = "ironhead-bucket" // 私有读
//var Bucket = "donscoco-bucket" // 公共读

// 注意：1.设置一下过期时间。

func main() {

	var futureDate = time.Date(2049, time.January, 10, 23, 0, 0, 0, time.UTC)

	fmt.Println(futureDate)

	oss.InitOSS(Endpoint, AccessKeyId, AccessKeySecret)

	file, err := os.Open("/Users/donscocochen/Pictures/resources/50x50.jpg")
	if err != nil {
		panic(err)
	}

	oss.PutObject(oss.DefaultOSS, Bucket, "tmp/50x50.jpg", file, time.Now().Add(100*time.Second), true)

	output, err := oss.GetObject(oss.DefaultOSS, Bucket, "tmp/50x50.jpg")
	if err != nil {
	}
	var buf [2048]byte
	n, err := output.Read(buf[:])
	fmt.Println(n)
	os.WriteFile("/Users/donscocochen/Pictures/50x50.jpg", buf[:], 0666)

	err = oss.DeleteObject(oss.DefaultOSS, Bucket, "tmp/580x326.png")
	if err != nil {
		panic(err)
	}

	//// 创建OSSClient实例。
	//// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为https://oss-cn-hangzhou.aliyuncs.com。其它Region请按实际情况填写。
	//// 阿里云账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM用户进行API访问或日常运维，请登录RAM控制台创建RAM用户。
	//client, err := oss.New(Endpoint, AccessKeyId, AccessKeySecret)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//
	//// 填写存储空间名称，例如examplebucket。
	//bucket, err := client.Bucket(Bucket)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//
	////////////// 本地文件上传 ///////////
	//// 依次填写Object的完整路径（例如exampledir/exampleobject.txt）和本地文件的完整路径（例如D:\\localpath\\examplefile.txt）。
	//err = bucket.PutObjectFromFile("test/580x326.png", "/Users/donscocochen/Pictures/580x326.png")
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	os.Exit(-1)
	//}
	//
	////////////// 数据流上传 ///////////
	////err = bucket.PutObject(objectname, io.Reader(res.Body))
	////if err != nil {
	////	fmt.Println("Error:", err)
	////	os.Exit(-1)
	////}
	//////////////

}

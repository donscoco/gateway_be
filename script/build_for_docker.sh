

GOPROXY=https://goproxy.cn
go env
# 预定义变量
docker build --build-arg IRONHEAD_PWD=${IRONHEAD_PWD} -t donscoco/gateway:v1 -f deploy/manager/Dockerfile .



## 先下载好包，待会docker化时放到镜像里面
go mod vendor
## docker login
docker build -t ${image}:${imageTag} deploy/debugweb/Dockerfile .
docker push ${image}:${imageTag}
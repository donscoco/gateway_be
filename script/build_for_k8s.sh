

GOPROXY=https://goproxy.cn
go env
# 预定义变量
PROJ_PATH=
DEPLOY_PATH=$PROJ_PATH/deploy
k8sDir=$(pwd)/deploy
image=donscoco/gateway_be
imageTag=v1

## 先下载好包，待会docker化时放到镜像里面
go mod vendor
## docker login
docker build -t ${image}:${imageTag} .
docker push ${image}:${imageTag}

#替换 版本宏
#grep '<IMAGE>' -rl ${k8sDir} | xargs sed -i 's#<IMAGE>#'"${image}"'#'
#grep '<IMAGE_TAG>' -rl ${k8sDir} | xargs sed -i 's#<IMAGE_TAG>#'"${imageTag}"'#'

kubectl apply -f ${k8sDir} ## 整个文件夹里面的yaml一起发布



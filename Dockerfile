FROM golang
MAINTAINER donscoco
WORKDIR /go/src/gateway_be
COPY . /go/src/gateway_be/
EXPOSE 30080
CMD ["/bin/bash", "/go/src/gateway_be/script/build_in_docker.sh"]

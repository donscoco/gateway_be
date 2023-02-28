FROM golang
MAINTAINER donscoco
WORKDIR /go/src/gateway_be
COPY . /go/src/gateway_be/

ARG IRONHEAD_PWD=deault_value
ENV IRONHEAD_PWD=$IRONHEAD_PWD

ARG GATEWAY_ENV=deault_value
ENV GATEWAY_ENV=$GATEWAY_ENV

EXPOSE 8701

RUN /bin/bash /go/src/gateway_be/script/build_in_docker.sh
CMD ["/go/src/gateway_be/bin/gateway_be"]

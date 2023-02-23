package grpc_proxy_middleware

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"google.golang.org/grpc"
	"log"
)

func GrpcFlowCountMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		totalCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowTotal)
		if err != nil {
			return err
		}
		totalCounter.Increase()
		serviceCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			return err
		}
		serviceCounter.Increase()

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcFlowCountMiddleware failed with error %v\n", err)
			return err
		}
		return nil
	}
}

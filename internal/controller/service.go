package controller

import (
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/donscoco/gateway_be/tool"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ServiceController struct{}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.GET("/service_detail", service.ServiceDetail)
	group.GET("/service_stat", service.ServiceStat)

	group.POST("/service_add_http", service.ServiceAddHTTP)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)

	group.POST("/service_add_tcp", service.ServiceAddTcp)
	group.POST("/service_update_tcp", service.ServiceUpdateTcp)

	group.POST("/service_add_grpc", service.ServiceAddGrpc)
	group.POST("/service_update_grpc", service.ServiceUpdateGrpc)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} bl.Response{data=model.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {
	params := &model.ServiceListInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	conf := iron_config.Conf

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//从db中分页读取基本信息
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//格式化输出信息
	outList := []model.ServiceListItemOutput{}
	for _, listItem := range list {
		serviceDetail, err := listItem.ServiceDetail(c, tx, &listItem)
		if err != nil {
			bl.ResponseError(c, 2003, err)
			return
		}
		//1、http后缀接入 clusterIP+clusterPort+path
		//2、http域名接入 domain
		//3、tcp、grpc接入 clusterIP+servicePort
		serviceAddr := "unknow"
		clusterIP := conf.GetString("/cluster/cluster_ip")
		// todo 如果设置成集群，改成拿到部署代理服务的节点的ip，例如负责部署的机器注册到zk，这边在zk中随机拿出来一个ip发请求，让其部署部署。
		// 现在是单节点，就先默认获取本机的ip
		if len(clusterIP) == 0 {
			clusterIP = tool.GetIpv4_192_168() // todo 这里不一定是获取 192.168,测试环境先这样
		}
		clusterPort := conf.GetString("/cluster/cluster_port")
		clusterSSLPort := conf.GetString("/cluster/cluster_ssl_port")
		if serviceDetail.Info.LoadType == bl.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == bl.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == bl.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == bl.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == bl.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == bl.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		if serviceDetail.Info.LoadType == bl.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType == bl.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		counter, err := counter.FlowCounterHandler.GetCounter(bl.FlowServicePrefix + listItem.ServiceName)
		if err != nil {
			bl.ResponseError(c, 2004, err)
			return
		}
		outItem := model.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}
	out := &model.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	bl.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_delete [get]
func (service *ServiceController) ServiceDelete(c *gin.Context) {
	params := &model.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save(c, tx); err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	bl.ResponseSuccess(c, "")
}

// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} bl.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (service *ServiceController) ServiceDetail(c *gin.Context) {
	params := &model.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	bl.ResponseSuccess(c, serviceDetail)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} bl.Response{data=model.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	params := &model.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	//读取基本信息
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}

	counter, err := counter.FlowCounterHandler.GetCounter(bl.FlowServicePrefix + serviceDetail.Info.ServiceName)
	if err != nil {
		bl.ResponseError(c, 2004, err)
		return
	}
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, time.UTC)
		hourData, _ := counter.GetHourData(dateTime, time.UTC)
		todayList = append(todayList, hourData)
	}

	yesterdayList := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, time.UTC)
		hourData, _ := counter.GetHourData(dateTime, time.UTC)
		yesterdayList = append(yesterdayList, hourData)
	}
	bl.ResponseSuccess(c, &model.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body model.ServiceAddHTTPInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (service *ServiceController) ServiceAddHTTP(c *gin.Context) {
	params := &model.ServiceAddHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if obj, err := serviceInfo.Find(c, tx, serviceInfo); err == nil && len(obj.ServiceName) != 0 { // 说明找到一个已经有相同名字的服务
		tx.Rollback()
		bl.ResponseError(c, 2002, errors.New("服务已存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if obj, err := httpUrl.Find(c, tx, httpUrl); err == nil && len(obj.Rule) != 0 {
		tx.Rollback()
		bl.ResponseError(c, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2005, err)
		return
	}
	//serviceModel.ID
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body model.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	params := &model.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2005, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2007, err)
		return
	}

	loadbalance := serviceDetail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
}

// ServiceAddHttp godoc
// @Summary tcp服务添加
// @Description tcp服务添加
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body model.ServiceAddTcpInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (admin *ServiceController) ServiceAddTcp(c *gin.Context) {
	params := &model.ServiceAddTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if obj, err := infoSearch.Find(c, tx, infoSearch); err == nil && len(obj.ServiceName) != 0 {
		bl.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if tcpRuleObj, err := tcpRuleSearch.Find(c, tx, tcpRuleSearch); err == nil && tcpRuleObj.Port != 0 {
		bl.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if grpcRuleObj, err := grpcRuleSearch.Find(c, tx, grpcRuleSearch); err == nil && grpcRuleObj.Port != 0 {
		bl.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx = tx.Begin()
	info := &dao.ServiceInfo{
		LoadType:    bl.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2007, err)
		return
	}

	httpRule := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary tcp服务更新
// @Description tcp服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body model.ServiceUpdateTcpInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (admin *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &model.ServiceUpdateTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	tx = tx.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.ServiceDetail(c, tx, service)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2003, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2004, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TCPRule != nil {
		tcpRule = detail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2005, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
	return
}

// ServiceAddHttp godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body model.ServiceAddGrpcInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (admin *ServiceController) ServiceAddGrpc(c *gin.Context) {
	params := &model.ServiceAddGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if obj, err := infoSearch.Find(c, tx, infoSearch); err == nil && len(obj.ServiceName) != 0 {
		bl.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if tcpRuleObj, err := tcpRuleSearch.Find(c, tx, tcpRuleSearch); err == nil && tcpRuleObj.Port != 0 {
		bl.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if grpcRuleObj, err := grpcRuleSearch.Find(c, tx, grpcRuleSearch); err == nil && grpcRuleObj.Port != 0 {
		bl.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx = tx.Begin()
	info := &dao.ServiceInfo{
		LoadType:    bl.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body model.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (admin *ServiceController) ServiceUpdateGrpc(c *gin.Context) {
	params := &model.ServiceUpdateGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		bl.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	tx = tx.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.ServiceDetail(c, tx, service)
	if err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2004, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcRule = detail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		bl.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	bl.ResponseSuccess(c, "")
	return
}

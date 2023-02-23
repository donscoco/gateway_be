package controller

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/donscoco/gateway_be/tool"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

// APPControllerRegister admin路由注册
func APPRegister(router *gin.RouterGroup) {
	admin := APPController{}
	router.GET("/app_list", admin.APPList)
	router.GET("/app_detail", admin.APPDetail)
	router.GET("/app_stat", admin.AppStatistics)
	router.GET("/app_delete", admin.APPDelete)
	router.POST("/app_add", admin.AppAdd)
	router.POST("/app_update", admin.AppUpdate)
}

type APPController struct {
}

// APPList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} bl.Response{data=model.APPListOutput} "success"
// @Router /app/app_list [get]
func (admin *APPController) APPList(c *gin.Context) {
	params := &model.APPListInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	info := &dao.App{}
	list, total, err := info.APPList(c, tx, params)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	outputList := []model.APPListItemOutput{}
	for _, item := range list {
		appCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowAppPrefix + item.AppID)
		if err != nil {
			bl.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outputList = append(outputList, model.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := model.APPListOutput{
		List:  outputList,
		Total: total,
	}
	bl.ResponseSuccess(c, output)
	return
}

// APPDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} bl.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (admin *APPController) APPDetail(c *gin.Context) {
	params := &model.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	// 获取连接池的一个连接
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	search := &dao.App{
		ID: params.ID,
	}
	detail, err := search.Find(c, tx, search)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	bl.ResponseSuccess(c, detail)
	return
}

// APPDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /app/app_delete [get]
func (admin *APPController) APPDelete(c *gin.Context) {
	params := &model.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	// 获取连接池的一个连接
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, tx, search)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	info.IsDelete = 1
	if err := info.Save(c, tx); err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	bl.ResponseSuccess(c, "")
	return
}

// AppAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param body body model.APPAddHttpInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /app/app_add [post]
func (admin *APPController) AppAdd(c *gin.Context) {
	params := &model.APPAddHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	// 获取连接池的一个连接
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, tx, search); err == nil {
		bl.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = tool.MD5(params.AppID)
	}
	info := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	bl.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body model.APPUpdateHttpInput true "body"
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /app/app_update [post]
func (admin *APPController) AppUpdate(c *gin.Context) {
	params := &model.APPUpdateHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	// 获取连接池的一个连接
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, tx, search)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = tool.MD5(params.AppID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, tx); err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	bl.ResponseSuccess(c, "")
	return
}

// AppStatistics godoc
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} bl.Response{data=model.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (admin *APPController) AppStatistics(c *gin.Context) {
	params := &model.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	// 获取连接池的一个连接
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	search := &dao.App{
		ID: params.ID,
	}
	detail, err := search.Find(c, tx, search)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	counter, err := counter.FlowCounterHandler.GetCounter(bl.FlowAppPrefix + detail.AppID)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		c.Abort()
		return
	}
	currentTime := time.Now()
	for i := 0; i <= time.Now().In(time.UTC).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, time.UTC)
		hourData, _ := counter.GetHourData(dateTime, time.UTC)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, time.UTC)
		hourData, _ := counter.GetHourData(dateTime, time.UTC)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := model.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	bl.ResponseSuccess(c, stat)
	return
}

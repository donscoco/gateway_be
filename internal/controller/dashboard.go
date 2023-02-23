package controller

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type DashboardController struct{}

func DashboardRegister(group *gin.RouterGroup) {
	service := &DashboardController{}
	group.GET("/panel_group_data", service.PanelGroupData)
	group.GET("/flow_stat", service.FlowStat)
	group.GET("/service_stat", service.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} bl.Response{data=model.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (service *DashboardController) PanelGroupData(c *gin.Context) {
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	_, serviceNum, err := serviceInfo.PageList(c, tx, &model.ServiceListInput{PageSize: 1, PageNo: 1})
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	app := &dao.App{}
	_, appNum, err := app.APPList(c, tx, &model.APPListInput{PageNo: 1, PageSize: 1})
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	counter, err := counter.FlowCounterHandler.GetCounter(bl.FlowTotal)
	if err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	out := &model.PanelGroupDataOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		TodayRequestNum: counter.TotalCount,
		CurrentQPS:      counter.QPS,
	}
	bl.ResponseSuccess(c, out)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} bl.Response{data=model.DashServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (service *DashboardController) ServiceStat(c *gin.Context) {
	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}
	legend := []string{}
	for index, item := range list {
		name, ok := bl.LoadTypeMap[item.LoadType]
		if !ok {
			bl.ResponseError(c, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}
	out := &model.DashServiceStatOutput{
		Legend: legend,
		Data:   list,
	}
	bl.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} bl.Response{data=model.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (service *DashboardController) FlowStat(c *gin.Context) {
	counter, err := counter.FlowCounterHandler.GetCounter(bl.FlowTotal)
	if err != nil {
		bl.ResponseError(c, 2001, err)
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

package service_manager

import (
	"errors"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
)

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap   map[string]*dao.ServiceDetail
	ServiceSlice []*dao.ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*dao.ServiceDetail{},
		ServiceSlice: []*dao.ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) GetTcpServiceList() []*dao.ServiceDetail {
	list := []*dao.ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == bl.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*dao.ServiceDetail {
	list := []*dao.ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == bl.LoadTypeGRPC {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*dao.ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != bl.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == bl.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == bl.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() {
		serviceInfo := &dao.ServiceInfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := gorm.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &model.ServiceListInput{PageNo: 1, PageSize: 99999}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			//fmt.Println("serviceDetail")
			//fmt.Println(tool.Obj2Json(serviceDetail))
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[listItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err
}

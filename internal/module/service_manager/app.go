package service_manager

import (
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"sync"
)

var AppManagerHandler *AppManager

func init() {
	AppManagerHandler = NewAppManager()
}

type AppManager struct {
	AppMap   map[string]*dao.App
	AppSlice []*dao.App
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*dao.App{},
		AppSlice: []*dao.App{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

func (s *AppManager) GetAppList() []*dao.App {
	return s.AppSlice
}

// todo domark
func (s *AppManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &dao.App{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := gorm.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &model.APPListInput{PageNo: 1, PageSize: 1000}
		list, _, err := appInfo.APPList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppID] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}

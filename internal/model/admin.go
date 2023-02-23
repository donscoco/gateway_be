package model

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type ChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"` //密码
}

func (param *ChangePwdInput) BindValiam(c *gin.Context) error {
	return bl.DefaultGetValidParams(c, param)
}

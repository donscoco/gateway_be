package controller

import (
	"encoding/json"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct{}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/logout", adminLogin.AdminLoginOut)
}

// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body model.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=model.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminlogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &model.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	// 获取连接池的一个连接

	tx, err := gorm.GetGormPool("default")
	if err != nil {
		bl.ResponseError(c, 2001, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, tx, params)
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//设置session
	sessInfo := &model.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessBts, err := json.Marshal(sessInfo)
	if err != nil {
		bl.ResponseError(c, 2003, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set(bl.AdminSessionInfoKey, string(sessBts))
	sess.Save()

	out := &model.AdminLoginOutput{
		Token: admin.UserName,
	}
	bl.ResponseSuccess(c, out)
}

// AdminLogin godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} bl.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminlogin *AdminLoginController) AdminLoginOut(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(bl.AdminSessionInfoKey)
	sess.Save()
	bl.ResponseSuccess(c, "")
}

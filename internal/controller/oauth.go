package controller

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/model"
	"github.com/donscoco/gateway_be/internal/module/service_manager"
	"github.com/donscoco/gateway_be/tool"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type OAuthController struct{}

func OAuthRegister(group *gin.RouterGroup) {
	oauth := &OAuthController{}
	group.POST("/tokens", oauth.Tokens)
}

// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body model.TokensInput true "body"
// @Success 200 {object} bl.Response{data=model.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauth *OAuthController) Tokens(c *gin.Context) {
	/*
		curl -i -X POST \
		   -H "Content-Type:application/json" \
		   -H "Authorization:Basic YXBwX2lkX2I6OGQ3YjExZWM5YmUwZTU5YTM2YjUyZjMyMzY2YzA5Y2I=" \
		   -d \
		'{
		  "grant_type": "client_credentials",
		  "scope": "read_write"
		}' \
		 'http://127.0.0.1:8080/oauth/tokens'
	*/
	params := &model.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		bl.ResponseError(c, 2000, err)
		return
	}

	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		bl.ResponseError(c, 2001, errors.New("用户名或密码格式错误"))
		return
	}

	//appSecret   = app_id_b:8d7b11ec9be0e59a36b52f32366c09cb
	appSecret, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		bl.ResponseError(c, 2002, err)
		return
	}

	//  取出 app_id secretß
	//  生成 app_list
	//  匹配 app_id
	//  基于 jwt生成token
	//  生成 output
	parts := strings.Split(string(appSecret), ":")
	if len(parts) != 2 {
		bl.ResponseError(c, 2003, errors.New("用户名或密码格式错误"))
		return
	}

	appList := service_manager.AppManagerHandler.GetAppList()
	for _, appInfo := range appList {
		if appInfo.AppID == parts[0] && appInfo.Secret == parts[1] {
			claims := jwt.StandardClaims{
				Issuer:    appInfo.AppID,
				ExpiresAt: time.Now().Add(bl.JwtExpires * time.Second).In(time.UTC).Unix(),
			}
			token, err := tool.JwtEncode(claims, bl.JwtSignKey)
			if err != nil {
				bl.ResponseError(c, 2004, err)
				return
			}
			output := &model.TokensOutput{
				ExpiresIn:   bl.JwtExpires,
				TokenType:   "Bearer",
				AccessToken: token,
				Scope:       "read_write",
			}
			bl.ResponseSuccess(c, output)
			return
		}
	}
	bl.ResponseError(c, 2005, errors.New("未匹配正确APP信息"))
}

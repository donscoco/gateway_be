package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

/*

1.JWT 的最大缺点是，由于服务器不保存 session 状态，因此无法在使用过程中废止某个 token，或者更改 token 的权限。
也就是说，一旦 JWT 签发了，在到期之前就会始终有效，除非服务器部署额外的逻辑。(加个版本号)

2.JWT 本身包含了认证信息，一旦泄露，任何人都可以获得该令牌的所有权限。
为了减少盗用，JWT 的有效期应该设置得比较短。对于一些比较重要的权限，使用时应该再次对用户进行认证。

*/

const (
	SignUser = "ironhead"
	SignKey  = "my_sign_key" // 加密用的秘钥，不能泄漏
)

type MyCustomClaims struct {
	Msg string `json:"msg,omitempty"` //自定义字段
	jwt.StandardClaims

	// jwt.StandardClaims 字段:
	//Audience  string `json:"aud,omitempty"`	//受众
	//ExpiresAt int64  `json:"exp,omitempty"`	//过期时间
	//Id        string `json:"jti,omitempty"`	//编号
	//IssuedAt  int64  `json:"iat,omitempty"`	//签发时间
	//Issuer    string `json:"iss,omitempty"`	//签发人
	//NotBefore int64  `json:"nbf,omitempty"`	//生效时间
	//Subject   string `json:"sub,omitempty"`	//主题
}

func Encode(msg string) (string, error) {
	mySigningKey := []byte(SignKey)
	claims := MyCustomClaims{
		msg,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 20).Unix(),
			Issuer:    SignUser,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

func Decode(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SignKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok {
		if claims.StandardClaims.ExpiresAt < time.Now().Unix() {
			return "", errors.New("request expired")
		}
		return claims.Msg, nil
	} else {
		return "", err
	}
}

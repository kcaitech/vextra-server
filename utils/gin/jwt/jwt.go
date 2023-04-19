package jwt

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/utils/jwt"
	"time"
)

type Data struct {
	Id       uint   `json:"id"`
	Nickname string `json:"nickname"`
}

var jwtSecret string = "123456"

func CreateJwt(jwtData *Data) (string, error) {
	t, err := jwt.NewJwt(jwt.NewHS256Signer(jwtSecret))
	if err != nil {
		return "", err
	}
	t.AddData("data", *jwtData)
	now := time.Now()
	t.SetRegisteredClaims(jwt.Payload{
		Exp: now.Add(time.Hour * 24 * 7).Unix(), // 过期时间
		Nbf: now.Unix(),                         // 生效时间
	})
	token, err := t.General()
	if err != nil {
		return "", err
	}
	return token, nil
}

func ParseJwt(token string) (*Data, error) {
	t, _ := jwt.NewJwt(jwt.NewHS256Signer(jwtSecret))
	payload, err := t.Parse(token)
	if err != nil {
		return nil, err
	}
	data, ok := payload["data"]
	if !ok {
		return nil, errors.New("无效token")
	}

	jsonBytes, _ := json.Marshal(data)
	var jwtData Data
	err = json.Unmarshal(jsonBytes, &jwtData)
	if err != nil {
		return nil, errors.New("无效token")
	}
	return &jwtData, nil
}

func GetJwtData(c *gin.Context) (*Data, error) {
	token := c.GetHeader("Token")
	return ParseJwt(token)
}

func GetUserId(c *gin.Context) (uint, error) {
	jwtData, err := GetJwtData(c)
	if err != nil {
		return 0, err
	}
	return jwtData.Id, nil
}

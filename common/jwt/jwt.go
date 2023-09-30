package jwt

import (
	"encoding/json"
	"errors"
	"protodesign.cn/kcserver/common/jwt/config"
	"protodesign.cn/kcserver/utils/jwt"
	"strings"
	"time"
)

type Data struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
}

func Init(filePath string) {
	_ = config.LoadConfig(filePath)
}

func CreateJwt(jwtData *Data) (string, error) {
	t, err := jwt.NewJwt(jwt.NewHS256Signer(config.Config.Jwt.Secret))
	if err != nil {
		return "", err
	}
	t.AddData("data", *jwtData)
	now := time.Now()
	t.SetRegisteredClaims(jwt.Payload{
		Exp: now.Add(time.Hour * time.Duration(config.Config.Jwt.ExpireHour)).Unix(), // 过期时间
		Nbf: now.Add((-1) * time.Minute).Unix(),                                      // 生效时间
	})
	token, err := t.General()
	if err != nil {
		return "", err
	}
	return token, nil
}

func ParseJwt(token string) (*Data, error) {
	t, _ := jwt.NewJwt(jwt.NewHS256Signer(config.Config.Jwt.Secret))
	payload, err := t.Parse(token)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, err
		} else {
			return nil, errors.New("无效token")
		}
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

func GetJwtFromAuthorization(authorization string) string {
	if !strings.HasPrefix(authorization, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authorization, "Bearer ")
}

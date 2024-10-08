package jwt

import (
	"testing"

	"kcaitech.com/kcserver/common/jwt/config"
)

func TestMain(m *testing.M) {
	Init(&config.LoadConfig("config_test.yaml").Jwt)
	m.Run()
}

func TestCreateJwt(t *testing.T) {
	jwtData := &Data{
		Id:       "110678381707673600",
		Nickname: "Jfeng",
	}
	token, err := CreateJwt(jwtData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}

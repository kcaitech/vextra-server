package jwt

import (
	"testing"
)

func TestMain(m *testing.M) {
	Init("config_test.yaml")
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

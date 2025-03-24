package services

import (
	"kcaitech.com/kcserver/provider/auth"
)

var jwtClient *auth.JWTClient

func Init(AuthServerURL string) {
	jwtClient = auth.NewJWTClient(AuthServerURL)
}

func GetJWTClient() *auth.JWTClient {
	return jwtClient
}

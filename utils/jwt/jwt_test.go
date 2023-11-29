package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// HS256算法测试
func TestHS256(t *testing.T) {
	jwt, err := NewJwt(NewHS256Signer("123456"))
	if err != nil {
		t.Fatal(err)
	}
	jwt.UpdateData(map[string]any{
		"userName": "Jfeng",
	})
	jwt.AddData("age", 27)
	jwtString, err := jwt.General()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jwtString)
	jwtString = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE3MDExNjQyMTAsIm5iZiI6MTcwMTE2MzQ5MCwiaWF0IjoxNzAxMTY0MDkwLCJkYXRhIjp7ImRhdGEiOnsiaWQiOiIxMTA2NzgzODE3MDc2NzM2MDAiLCJuaWNrbmFtZSI6IkpmZW5nIn19fQ.-weSkp4etm8r5aDZMxQ1-EnMwgWB2UjuT7dcxOFQ5yE"
	jwtParseRes, err := jwt.ParsePayload(jwtString)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(jwtParseRes)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}

// RS256算法测试
func TestRS256(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	jwt, err := NewJwt(NewRS256Encryptor(privateKey, publicKey))
	if err != nil {
		t.Fatal(err)
	}
	jwt.UpdateData(map[string]any{
		"userName": "Jfeng",
	})
	jwt.AddData("age", 27)
	jwtString, err := jwt.General()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jwtString)
	jwtParseRes, err := jwt.Parse(jwtString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jwtParseRes)
}

// Payload中的标准声明测试（过期时间Exp、生效时间Nbf）
func TestRegisteredClaims(t *testing.T) {
	jwt, err := NewJwt(NewHS256Signer("123456"))
	if err != nil {
		t.Fatal(err)
	}
	jwt.UpdateData(map[string]any{
		"userName": "Jfeng",
	})
	jwt.AddData("age", 27)
	now := time.Now()
	jwt.SetRegisteredClaims(Payload{
		Exp: now.Add(time.Second * 1).Unix(),    // 1秒后过期
		Nbf: now.Add(time.Second * (-1)).Unix(), // 1秒前生效
	})
	jwtString, err := jwt.General()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jwtString)
	jwtParseRes, err := jwt.Parse(jwtString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(jwtParseRes)
}

func TestCreateJwt(t *testing.T) {
	type Data struct {
		Id       string `json:"id"`
		Nickname string `json:"nickname"`
	}
	jwtData := Data{
		Id:       "110678381707673600",
		Nickname: "Jfeng",
	}
	jwt, err := NewJwt(NewHS256Signer("123456"))
	if err != nil {
		t.Fatal(err)
	}
	jwt.AddData("data", jwtData)
	now := time.Now()
	jwt.SetRegisteredClaims(Payload{
		//Exp: now.Add(time.Hour * time.Duration(168)).Unix(), // 过期时间
		Exp: now.Add(time.Minute * time.Duration(12)).Unix(), // 过期时间
		//Nbf: now.Add((-60) * time.Minute).Unix(),             // 生效时间
	})
	token, err := jwt.General()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}

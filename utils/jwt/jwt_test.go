package jwt

import (
	"crypto/rand"
	"crypto/rsa"
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
	jwt.UpdateData(map[string]interface{}{
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
	jwt.UpdateData(map[string]interface{}{
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
	jwt.UpdateData(map[string]interface{}{
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

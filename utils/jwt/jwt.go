package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func ByteSliceEqual(s0 []byte, s1 []byte) bool {
	l := len(s0)
	if l != len(s1) {
		return false
	}
	for i := 0; i < l; i++ {
		if s0[i] != s1[i] {
			return false
		}
	}
	return true
}

// 签名器
type signer interface {
	algorithmName() string                       // 返回加密算法名称，需要对应Header部分中的Alg属性
	sign(data []byte) ([]byte, error)            // 签名
	verify(data []byte, encryptData []byte) bool // 验证
}

// HS256签名器
type hs256Signer struct {
	secret string
}

func NewHS256Signer(secret string) hs256Signer {
	return hs256Signer{
		secret: secret,
	}
}

func (that hs256Signer) algorithmName() string {
	return "HS256"
}

func (that hs256Signer) sign(data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(that.secret))
	h.Write(data)
	return h.Sum(nil), nil
}

func (that hs256Signer) verify(data []byte, encryptData []byte) bool {
	res, _ := that.sign(data)
	return ByteSliceEqual(res, encryptData)
}

// RS256签名器
type rs256Signer struct {
	privateKey *rsa.PrivateKey // 私钥，签名时必须
	publicKey  *rsa.PublicKey  // 公钥，验证时必须
}

func NewRS256Encryptor(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) rs256Signer {
	return rs256Signer{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (that rs256Signer) algorithmName() string {
	return "RS256"
}

func (that rs256Signer) sign(data []byte) ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	hashed := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, that.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (that rs256Signer) verify(data []byte, encryptData []byte) bool {
	hash := sha256.New()
	_, err := hash.Write(data)
	if err != nil {
		return false
	}
	hashed := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(that.publicKey, crypto.SHA256, hashed, encryptData)
	return err == nil
}

// JWT Header部分
type header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

// Payload JWT Payload部分
type Payload struct {
	Iss  string                 `json:"iss,omitempty"`  // 发行人
	Sub  string                 `json:"sub,omitempty"`  // 主题
	Aud  string                 `json:"aud,omitempty"`  // 用户
	Exp  int64                  `json:"exp,omitempty"`  // 过期时间（时间戳，秒）
	Nbf  int64                  `json:"nbf,omitempty"`  // 生效时间（时间戳，秒）
	Iat  int64                  `json:"iat"`            // 签发时间（时间戳，秒）
	Jti  string                 `json:"jti,omitempty"`  // jwt id
	Data map[string]interface{} `json:"data,omitempty"` // 自定义数据
}

type Jwt struct {
	header    header
	encryptor signer
	payload   Payload
	signature []byte

	_headerBase64    string
	_payloadBase64   string
	_payloadJson     string
	_signatureBase64 string

	jwtString string
}

func NewJwt(encryptor signer) (Jwt, error) {
	var alg string
	switch encryptor.(type) {
	case hs256Signer:
		alg = "HS256"
	case rs256Signer:
		alg = "RS256"
	default:
		return Jwt{}, errors.New("加密器类型不支持")
	}
	return Jwt{
		header: header{
			Typ: "JWT",
			Alg: alg,
		},
		encryptor: encryptor,
		payload: Payload{
			Iat:  time.Now().Unix(), // 签发时间
			Data: map[string]interface{}{},
		},
	}, nil
}

// SetRegisteredClaims 设置Payload中标准声明的部分属性
func (that *Jwt) SetRegisteredClaims(payload Payload) {
	if payload.Iss != "" {
		that.payload.Iss = payload.Iss
	}
	if payload.Sub != "" {
		that.payload.Sub = payload.Sub
	}
	if payload.Aud != "" {
		that.payload.Aud = payload.Aud
	}
	if payload.Exp != 0 {
		that.payload.Exp = payload.Exp
	}
	if payload.Nbf != 0 {
		that.payload.Nbf = payload.Nbf
	}
	if payload.Iat != 0 {
		that.payload.Iat = payload.Iat
	}
	if payload.Jti != "" {
		that.payload.Jti = payload.Jti
	}
}

// UpdateData 往Payload中添加多条数据（claim）
func (that *Jwt) UpdateData(data map[string]interface{}) {
	for k, v := range data {
		that.payload.Data[k] = v
	}
}

// AddData 往Payload中添加一条数据（claim）
func (that *Jwt) AddData(key string, value interface{}) {
	that.payload.Data[key] = value
}

// General 生成JWT
func (that *Jwt) General() (string, error) {
	var err error
	var temp []byte

	// Header
	temp, err = json.Marshal(that.header)
	if err != nil {
		return "", errors.New("header json编码失败")
	}
	that._headerBase64 = base64.RawURLEncoding.EncodeToString(temp)

	// Payload
	temp, err = json.Marshal(that.payload)
	if err != nil {
		return "", errors.New("payload json编码失败")
	}
	that._payloadJson = string(temp)
	that._payloadBase64 = base64.RawURLEncoding.EncodeToString(temp)

	// Signature
	that.signature, err = that.encryptor.sign([]byte(that._headerBase64 + "." + that._payloadBase64))
	if err != nil {
		return "", errors.New("签名失败")
	}
	that._signatureBase64 = base64.RawURLEncoding.EncodeToString(that.signature)
	if err != nil {
		return "", errors.New("signature base64编码失败")
	}

	that.jwtString = fmt.Sprintf("%s.%s.%s", that._headerBase64, that._payloadBase64, that._signatureBase64)

	return that.jwtString, nil
}

// Parse 验证并解析JWT，返回Payload中的自定义数据
func (that *Jwt) Parse(jwtString string) (res map[string]interface{}, err error) {
	// 分割
	splitRes := strings.Split(jwtString, ".")
	if len(splitRes) != 3 {
		return res, errors.New("token格式错误")
	}
	that._headerBase64, that._payloadBase64, that._signatureBase64 = splitRes[0], splitRes[1], splitRes[2]

	var temp []byte

	// Signature
	if that.signature, err = base64.RawURLEncoding.DecodeString(that._signatureBase64); err != nil {
		return res, errors.New("signature格式错误")
	}
	if !that.encryptor.verify([]byte(that._headerBase64+"."+that._payloadBase64), that.signature) {
		return res, errors.New("无效token")
	}

	// Header
	if temp, err = base64.RawURLEncoding.DecodeString(that._headerBase64); err != nil {
		return res, errors.New("header base64解码失败")
	}
	if err = json.Unmarshal(temp, &that.header); err != nil {
		return res, errors.New("header json解码失败")
	}
	if that.header.Typ != "JWT" || that.header.Alg != that.encryptor.algorithmName() {
		return res, errors.New("不支持的类型：" + that.header.Alg)
	}

	// Payload
	if temp, err = base64.RawURLEncoding.DecodeString(that._payloadBase64); err != nil {
		return res, errors.New("payload base64解码失败")
	}
	that._payloadJson = string(temp)
	if err = json.Unmarshal(temp, &that.payload); err != nil {
		return res, errors.New("payload json解码失败")
	}

	// 验证有效性
	now := time.Now().Unix()
	if that.payload.Exp != 0 && that.payload.Exp <= now {
		return res, errors.New("token已过期")
	}
	if that.payload.Nbf != 0 && that.payload.Nbf > now {
		return res, errors.New("token未生效")
	}

	return that.payload.Data, nil
}

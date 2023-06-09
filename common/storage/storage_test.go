package storage

import (
	"log"
	"protodesign.cn/kcserver/utils/storage/base"
	"testing"
)

func TestGenerateAccessKey(t *testing.T) {
	if err := Init("config_test.yaml"); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	key, err := Bucket.GenerateAccessKey("*", base.AuthOpGetObject, 3600*24*365)
	if err != nil {
		log.Fatalln("生成密钥失败" + err.Error())
	}
	log.Println(key)
}

func TestCopyDirectory(t *testing.T) {
	if err := Init("config_test.yaml"); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	if _, err := Bucket.CopyDirectory("0177361a-736f-4375-bd77-4ab1a0675fd9", "0000000000000000000"); err != nil {
		log.Fatalln("复制目录失败：" + err.Error())
	}
}

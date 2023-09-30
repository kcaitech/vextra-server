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
	key, err := Bucket.GenerateAccessKey("*", base.AuthOpGetObject, 3600*24*365, "testSessionName")
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

func TestGetObjectInfo(t *testing.T) {
	var err error
	if err = Init("config_test.yaml"); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	var documentInfo *base.ObjectInfo
	if documentInfo, err = Bucket.GetObjectInfo("9b4482f8-c1fd-47ce-b19b-2387b963f3f9/document-meta.json"); err != nil {
		log.Fatalln("获取文件信息失败：" + err.Error())
	}
	log.Println(documentInfo.VersionID)
}

package storage

import (
	"log"
	"os"
	"kcaitech.com/kcserver/utils/storage/base"
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
	if _, err := Bucket.CopyDirectory("0177361a-736f-4375-bd77-4ab1a0675fd9", "1"); err != nil {
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

func TestPutObject(t *testing.T) {
	var err error
	if err = Init("config_test.yaml"); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	// 读取本地的test_image.jpg文件
	file, err := os.Open("test_image.jpg")
	// 获取文件大小
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	if _, err := FilesBucket.PutObject(&base.PutObjectInput{
		ObjectName:  "/users/cbf9f052-1e92-4b2c-a899-12d8a5f47369/avatar/af93cb99-a945-413e-a099-e69d67e3ecf3.jpg",
		Reader:      file,
		ObjectSize:  fileSize,
		ContentType: "image/jpeg",
	}); err != nil {
		log.Fatalln("上传文件失败：" + err.Error())
	}
}

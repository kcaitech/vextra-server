package safereview

import (
	"log"
	"protodesign.cn/kcserver/common/storage"
	"testing"
)

func TestMain(m *testing.M) {
	if err := storage.Init("../storage/config_test.yaml"); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	if err := Init("config_test.yaml"); err != nil {
		log.Fatalln("safereview初始化失败：" + err.Error())
	}
	m.Run()
}

func TestReviewText(t *testing.T) {
	response, err := Client.ReviewText("习近平不好")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}

func TestReviewPicture(t *testing.T) {
	response, err := Client.ReviewPicture(
		"https://img.alicdn.com/tfs/TB1U4r9AeH2gK0jSZJnXXaT1FXa-2880-480.png",
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}

func TestReviewPictureFromStorage(t *testing.T) {
	//bucketConfig := storage.Bucket.GetConfig()
	//response, err := Client.ReviewPictureFromStorage(bucketConfig.Region, bucketConfig.BucketName, "test.jpg")
	response, err := Client.ReviewPictureFromStorage(
		"cn-hangzhou",
		"protodesign-document",
		//"0000000000000000000/medias/TB1U4r9AeH2gK0jSZJnXXaT1FXa-2880-480.png",
		"0000000000000000000/medias/2560x1440.297.webp",
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}

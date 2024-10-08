package safereview

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"kcaitech.com/kcserver/common/safereview/config"
	"kcaitech.com/kcserver/common/storage"
	storageConf "kcaitech.com/kcserver/common/storage/config"
)

func TestMain(m *testing.M) {
	sconf := storageConf.LoadConfig("../storage/config_test.yaml")
	if err := storage.Init(&sconf.Storage); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	conf := config.LoadConfig("config_test.yaml")
	if err := Init(&conf.SafeReview); err != nil {
		log.Fatalln("safereview初始化失败：" + err.Error())
	}
	time.Sleep(1 * time.Second)
	m.Run()
}

func TestReviewText(t *testing.T) {
	log.Println("TestReviewText")
	response, err := Client.ReviewText("习近平不好 免费翻墙 电话 找小姐 张三 我是你爹 他妈的")
	if err != nil {
		t.Fatal(err)
	}
	log.Println(response.Status)
	log.Println(response.Labels)
	log.Println(response.Reason)
	log.Println(response.Words)
}

func TestReviewPicture(t *testing.T) {
	response, err := Client.ReviewPictureFromUrl(
		//"https://img.alicdn.com/tfs/TB1U4r9AeH2gK0jSZJnXXaT1FXa-2880-480.png",
		"https://storage1.kcaitech.com/0000000000000000000/medias/images.jpg?Expires=1701428203&OSSAccessKeyId=TMP.3KfXnUrcedsEhsiQMh14bseRCk6q8FK58yZ8MMsYiHk3HEYqsBV8mdX2vX4pZtrG2vh69SCCG4ANL2H4e6AaiUnFxpHJey&Signature=7mMlaP2yFut6U52%2FHHEMF3eZfPQ%3D",
	)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(response.Status)
	log.Println(response.Result)
}

//func TestReviewPictureFromStorage(t *testing.T) {
//	//bucketConfig := storage.Bucket.GetConfig()
//	//response, err := Client.ReviewPictureFromStorage(bucketConfig.Region, bucketConfig.BucketName, "test.jpg")
//	response, err := Client.ReviewPictureFromStorage(
//		"cn-hangzhou",
//		"protodesign-document",
//		//"0000000000000000000/medias/TB1U4r9AeH2gK0jSZJnXXaT1FXa-2880-480.png",
//		"0000000000000000000/medias/2560x1440.297.webp",
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//	log.Println(response.Status)
//	log.Println(response.Result)
//}

func TestReviewPictureFromBase64(t *testing.T) {
	file, err := os.Open("test_image_0.jpg")
	if err != nil {
		log.Println("打开文件失败", err)
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("读取文件失败", err)
		return
	}

	base64Str := base64.StdEncoding.EncodeToString(imgData)
	log.Println(base64Str[:100] + "...")
	response, err := Client.ReviewPictureFromBase64(base64Str)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(response.Status)
	log.Println(response.Result)
	log.Println(response.Reason)
}

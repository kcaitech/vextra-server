package safereview

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	myMath "kcaitech.com/kcserver/utils/math"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type BaiduClient struct {
	*censor.ContentCensorClient
}

type TextResponse struct {
	LogId          int    `json:"log_id"`
	ErrorCode      *int   `json:"error_code"`
	ErrorMsg       string `json:"error_msg"`
	ConclusionType *int   `json:"conclusionType"`
	Conclusion     string `json:"conclusion"`
	Data           []struct {
		ErrorCode *int   `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		Type      *int   `json:"type"`
		SubType   *int   `json:"subType"`
		Msg       string `json:"msg"`
		Hits      []struct {
			Probability *float64 `json:"probability"`
			DatasetName string   `json:"datasetName"`
			Words       []string `json:"words"`
		} `json:"hits"`
	} `json:"data"`
}

func (c *BaiduClient) reviewText0(text string) (*ReviewTextResponse, error) {
	resStr := c.TextCensor(text)
	res := TextResponse{}
	if err := json.Unmarshal([]byte(resStr), &res); err != nil {
		return nil, err
	}
	if res.ConclusionType == nil && res.ErrorCode == nil {
		return nil, errors.New("error_code和conclusionType都为空")
	}
	if res.ErrorCode != nil {
		return nil, errors.New(fmt.Sprintf("接口调用失败 errorCode：%s errorMsg：%s", res.ErrorCode, res.ErrorMsg))
	}

	status := ReviewTextResultPass
	if *res.ConclusionType == 3 {
		status = ReviewTextResultSuspected
	} else if *res.ConclusionType != 1 {
		status = ReviewTextResultBlock
	}

	reasons := []string{}
	labels := []string{}
	words := []string{}
	for _, data := range res.Data {
		if data.ErrorCode != nil || !(data.Type != nil && data.SubType != nil) {
			continue
		}
		reasons = append(reasons, data.Msg)
		for _, hit := range data.Hits {
			labels = append(labels, hit.DatasetName)
			words = append(words, hit.Words...)
		}
	}
	reasons = sliceutil.Unique(nil, reasons...)
	labels = sliceutil.Unique(nil, labels...)
	words = sliceutil.Unique(nil, words...)

	return &ReviewTextResponse{
		Status: status,
		Reason: strings.Join(reasons, ","),
		Labels: labels,
		Words:  words,
	}, nil
}

func (c *BaiduClient) reviewText(text string) (*ReviewTextResponse, error) {
	for i := 0; i < 10; i++ {
		res, err := c.reviewText0(text)
		if err == nil {
			return res, nil
		} else if strings.Contains(err.Error(), "qps request limit") {
			time.Sleep(time.Duration(400+rand.Intn(200)) * time.Millisecond) // 等待500+-100ms
		} else {
			return nil, err
		}
	}
	return nil, errors.New("qps request limit")
}

func (c *BaiduClient) ReviewText(text string) (*ReviewTextResponse, error) {
	if len([]rune(text)) <= 5000 {
		return c.reviewText(text)
	}
	textRuneList := []rune(text)
	response, err := c.reviewText(string(textRuneList[0:5000]))
	if err != nil {
		return nil, err
	}
	count := myMath.IntDivideCeil(len(textRuneList)-100, 4900)
	for i := 1; i < count; i++ {
		start := i * 4900
		end := i*4900 + 5000
		if i == count-1 {
			end = len(textRuneList)
		}
		partResponse, err := c.reviewText(string(textRuneList[start:end]))
		if err != nil {
			return nil, err
		}
		if partResponse.Status == ReviewTextResultPass {
			continue
		}
		if response.Status == ReviewTextResultPass {
			response = partResponse
		}
		response.Reason = response.Reason + "," + partResponse.Reason
		response.Labels = append(response.Labels, partResponse.Labels...)
	}
	return response, nil
}

type ImageResponse struct {
	LogId          int    `json:"log_id"`
	ErrorCode      *int   `json:"error_code"`
	ErrorMsg       string `json:"error_msg"`
	ConclusionType *int   `json:"conclusionType"`
	Conclusion     string `json:"conclusion"`
	Data           []struct {
		ErrorCode   *int     `json:"error_code"`
		ErrorMsg    string   `json:"error_msg"`
		Type        *int     `json:"type"`
		SubType     *int     `json:"subType"`
		Msg         string   `json:"msg"`
		Probability *float64 `json:"probability"`
		Hits        []struct {
			Words []string `json:"words"`
		} `json:"hits"`
	} `json:"data"`
}

func (c *BaiduClient) reviewPictureParse(resStr string) (*ReviewImageResponse, error) {
	res := ImageResponse{}
	if err := json.Unmarshal([]byte(resStr), &res); err != nil {
		return nil, err
	}
	if res.ConclusionType == nil && res.ErrorCode == nil {
		return nil, errors.New("error_code和conclusionType都为空")
	}
	if res.ErrorCode != nil {
		return nil, errors.New(fmt.Sprintf("接口调用失败 errorCode：%s errorMsg：%s", res.ErrorCode, res.ErrorMsg))
	}

	status := ReviewImageResultPass
	if *res.ConclusionType == 3 {
		status = ReviewImageResultSuspected
	} else if *res.ConclusionType != 1 {
		status = ReviewImageResultBlock
	}

	results := []ReviewImageResultItem{}
	for _, data := range res.Data {
		if data.ErrorCode != nil || !(data.Type != nil && data.SubType != nil) {
			continue
		}
		probability := float64(0)
		if data.Probability != nil {
			probability = *data.Probability
		}
		results = append(results, ReviewImageResultItem{
			Reason:     data.Msg,
			Confidence: probability,
		})
	}
	reason := strings.Join(
		sliceutil.MapT(func(item ReviewImageResultItem) string {
			return item.Reason
		}, results...),
		",",
	)

	return &ReviewImageResponse{
		Status: status,
		Result: results,
		Reason: reason,
	}, nil
}

type reviewPictureType uint8

const (
	reviewPictureTypeUrl reviewPictureType = iota
	reviewPictureTypeBase64
)

func (c *BaiduClient) reviewPicture(t reviewPictureType, imageValue string) (*ReviewImageResponse, error) {
	var imgCensorFunc func(imgUrl string, options map[string]interface{}) (result string)
	if t == reviewPictureTypeUrl {
		imgCensorFunc = c.ImgCensorUrl
	} else if t == reviewPictureTypeBase64 {
		imgCensorFunc = c.ImgCensor
	}
	for i := 0; i < 10; i++ {
		res, err := c.reviewPictureParse(imgCensorFunc(imageValue, nil))
		if err == nil {
			return res, nil
		} else if strings.Contains(err.Error(), "qps request limit") {
			time.Sleep(time.Duration(400+rand.Intn(200)) * time.Millisecond) // 等待500+-100ms
		} else {
			return nil, err
		}
	}
	return nil, errors.New("qps request limit")
}

func (c *BaiduClient) ReviewPictureFromUrl(imageUrl string) (*ReviewImageResponse, error) {
	return c.reviewPicture(reviewPictureTypeUrl, imageUrl)
}

func (c *BaiduClient) ReviewPictureFromBase64(imageBase64 string) (*ReviewImageResponse, error) {
	return c.reviewPicture(reviewPictureTypeBase64, imageBase64)
}

func NewBaiduClient(conf *SafeReviewConf) (*BaiduClient, error) {
	client := &BaiduClient{
		censor.NewClient(conf.Baidu.ApiKey, conf.Baidu.SecretKey),
	}
	return client, nil
}

// var Client base.Client

// func Init(conf *config.SafeReviewConf) error {
// 	// conf := config.LoadConfig(filePath)
// 	Client = &BaiduClient{
// 		censor.NewClient(conf.Baidu.ApiKey, conf.Baidu.SecretKey),
// 	}
// 	return nil
// }

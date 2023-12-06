package baidu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	"protodesign.cn/kcserver/common/safereview/base"
	"protodesign.cn/kcserver/common/safereview/config"
	myMath "protodesign.cn/kcserver/utils/math"
	"protodesign.cn/kcserver/utils/sliceutil"
	"strings"
)

type client struct {
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

func (c *client) reviewText(text string) (*base.ReviewTextResponse, error) {
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

	status := base.ReviewTextResultPass
	if *res.ConclusionType == 3 {
		status = base.ReviewTextResultSuspected
	} else if *res.ConclusionType != 1 {
		status = base.ReviewTextResultBlock
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

	return &base.ReviewTextResponse{
		Status: status,
		Reason: strings.Join(reasons, ","),
		Labels: labels,
		Words:  words,
	}, nil
}

func (c *client) ReviewText(text string) (*base.ReviewTextResponse, error) {
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
		if partResponse.Status == base.ReviewTextResultPass {
			continue
		}
		if response.Status == base.ReviewTextResultPass {
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

func (c *client) reviewPictureParse(resStr string) (*base.ReviewImageResponse, error) {
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

	status := base.ReviewImageResultPass
	if *res.ConclusionType == 3 {
		status = base.ReviewImageResultSuspected
	} else if *res.ConclusionType != 1 {
		status = base.ReviewImageResultBlock
	}

	results := []base.ReviewImageResultItem{}
	for _, data := range res.Data {
		if data.ErrorCode != nil || !(data.Type != nil && data.SubType != nil) {
			continue
		}
		probability := float64(0)
		if data.Probability != nil {
			probability = *data.Probability
		}
		results = append(results, base.ReviewImageResultItem{
			Reason:     data.Msg,
			Confidence: probability,
		})
	}
	reason := strings.Join(
		sliceutil.MapT(func(item base.ReviewImageResultItem) string {
			return item.Reason
		}, results...),
		",",
	)

	return &base.ReviewImageResponse{
		Status: status,
		Result: results,
		Reason: reason,
	}, nil
}

func (c *client) ReviewPictureFromUrl(imageUrl string) (*base.ReviewImageResponse, error) {
	return c.reviewPictureParse(c.ImgCensorUrl(imageUrl, nil))
}

func (c *client) ReviewPictureFromBase64(imageBase64 string) (*base.ReviewImageResponse, error) {
	return c.reviewPictureParse(c.ImgCensor(imageBase64, nil))
}

var Client base.Client

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)
	Client = &client{
		censor.NewClient(conf.Baidu.ApiKey, conf.Baidu.SecretKey),
	}
	return nil
}

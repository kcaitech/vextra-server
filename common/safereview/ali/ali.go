package ali

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/common/safereview/config"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type client struct {
	*green20220302.Client
}

func (c *client) ReviewText(text string) (*base.ReviewTextResponse, error) {
	serviceParameters, _ := json.Marshal(
		map[string]any{
			"content": text,
		},
	)

	textModerationRequest := &green20220302.TextModerationRequest{
		// 有效值参考：https://yundun.console.aliyun.com/?p=cts#/textReview/ruleConfig/cn-shanghai）
		Service:           tea.String("ad_compliance_detection"),
		ServiceParameters: tea.String(string(serviceParameters)),
	}
	response, err := c.TextModerationWithOptions(textModerationRequest, &util.RuntimeOptions{
		ReadTimeout:    tea.Int(10000),
		ConnectTimeout: tea.Int(10000),
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("response为nil")
	}
	statusCode := tea.IntValue(tea.ToInt(response.StatusCode))
	if statusCode != http.StatusOK {
		return nil, errors.New("response code异常：" + tea.ToString(statusCode))
	}
	body := response.Body
	bodyCode := body.Code
	bodyData := body.Data
	bodyMessage := body.Message
	if *bodyCode != 200 {
		return nil, errors.New("response body code异常：" + tea.ToString(*bodyCode) + "，message：" + tea.StringValue(bodyMessage))
	}

	reason := tea.StringValue(bodyData.Reason)
	var labels []string
	status := base.ReviewTextResultPass
	if len(reason) > 0 {
		status = base.ReviewTextResultBlock
		labels = strings.Split(tea.StringValue(bodyData.Labels), ",")
	}

	return &base.ReviewTextResponse{
		Status: status,
		Reason: reason,
		Labels: labels,
	}, nil
}

func (c *client) reviewPicture(serviceParameters string) (*base.ReviewImageResponse, error) {
	imageModerationRequest := &green20220302.ImageModerationRequest{
		// 有效值参考：https://yundun.console.aliyun.com/?p=cts#/pictureReview/ruleConfig/cn-shanghai）
		Service:           tea.String("baselineCheck"),
		ServiceParameters: tea.String(serviceParameters),
	}
	response, err := c.ImageModerationWithOptions(imageModerationRequest, &util.RuntimeOptions{
		ReadTimeout:    tea.Int(10000),
		ConnectTimeout: tea.Int(10000),
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("response为nil")
	}
	statusCode := tea.IntValue(tea.ToInt(response.StatusCode))
	if statusCode != http.StatusOK {
		return nil, errors.New("response code异常：" + tea.ToString(statusCode))
	}
	body := response.Body
	bodyCode := body.Code
	bodyData := body.Data
	bodyMessage := body.Msg
	if *bodyCode != 200 {
		return nil, errors.New("response body code异常：" + tea.ToString(*bodyCode) + "，message：" + tea.StringValue(bodyMessage))
	}

	results := sliceutil.MapT(func(item *green20220302.ImageModerationResponseBodyDataResult) base.ReviewImageResultItem {
		return base.ReviewImageResultItem{
			Reason:     tea.StringValue(item.Label),
			Confidence: float64(tea.Float32Value(item.Confidence)),
		}
	}, bodyData.Result...)
	results = sliceutil.FilterT(func(item base.ReviewImageResultItem) bool {
		return item.Reason != "nonLabel"
	}, results...)
	reason := strings.Join(
		sliceutil.MapT(func(item base.ReviewImageResultItem) string {
			return item.Reason
		}, results...),
		",",
	)

	status := base.ReviewImageResultPass
	if len(results) > 0 {
		status = base.ReviewImageResultBlock
	}

	return &base.ReviewImageResponse{
		Status: status,
		Result: results,
		Reason: reason,
	}, nil
}

func (c *client) ReviewPictureFromUrl(imageUrl string) (*base.ReviewImageResponse, error) {
	serviceParameters, _ := json.Marshal(
		map[string]any{
			"imageUrl": imageUrl,
		},
	)
	return c.reviewPicture(string(serviceParameters))
}

func (c *client) ReviewPictureFromStorage(regionName string, bucketName string, objectName string) (*base.ReviewImageResponse, error) {
	serviceParameters, _ := json.Marshal(
		map[string]any{
			"ossRegionId":   regionName,
			"ossBucketName": bucketName,
			"ossObjectName": objectName,
		},
	)
	return c.reviewPicture(string(serviceParameters))
}

func (c *client) ReviewPictureFromBase64(imageBase64 string) (*base.ReviewImageResponse, error) {
	return nil, errors.New("图片审核接口不支持base64格式")
}

var Client base.Client

func Init(conf *config.SafeReviewConf) error {
	// conf := config.LoadConfig(filePath)
	_client, err := green20220302.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(conf.Ali.AccessKeyId),
		AccessKeySecret: tea.String(conf.Ali.AccessKeySecret),
		RegionId:        tea.String(conf.Ali.RegionId),
		Endpoint:        tea.String(conf.Ali.Endpoint),
		ConnectTimeout:  tea.Int(3000),
		ReadTimeout:     tea.Int(6000),
	})
	if err != nil {
		return err
	}
	Client = &client{_client}
	return nil
}

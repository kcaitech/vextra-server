package safereview

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type AliClient struct {
	*green20220302.Client
}

func (c *AliClient) ReviewText(text string) (*ReviewTextResponse, error) {
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
	status := ReviewTextResultPass
	if len(reason) > 0 {
		status = ReviewTextResultBlock
		labels = strings.Split(tea.StringValue(bodyData.Labels), ",")
	}

	return &ReviewTextResponse{
		Status: status,
		Reason: reason,
		Labels: labels,
	}, nil
}

func (c *AliClient) reviewPicture(serviceParameters string) (*ReviewImageResponse, error) {
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

	results := sliceutil.MapT(func(item *green20220302.ImageModerationResponseBodyDataResult) ReviewImageResultItem {
		return ReviewImageResultItem{
			Reason:     tea.StringValue(item.Label),
			Confidence: float64(tea.Float32Value(item.Confidence)),
		}
	}, bodyData.Result...)
	results = sliceutil.FilterT(func(item ReviewImageResultItem) bool {
		return item.Reason != "nonLabel"
	}, results...)
	reason := strings.Join(
		sliceutil.MapT(func(item ReviewImageResultItem) string {
			return item.Reason
		}, results...),
		",",
	)

	status := ReviewImageResultPass
	if len(results) > 0 {
		status = ReviewImageResultBlock
	}

	return &ReviewImageResponse{
		Status: status,
		Result: results,
		Reason: reason,
	}, nil
}

func (c *AliClient) ReviewPictureFromUrl(imageUrl string) (*ReviewImageResponse, error) {
	serviceParameters, _ := json.Marshal(
		map[string]any{
			"imageUrl": imageUrl,
		},
	)
	return c.reviewPicture(string(serviceParameters))
}

func (c *AliClient) ReviewPictureFromStorage(regionName string, bucketName string, objectName string) (*ReviewImageResponse, error) {
	serviceParameters, _ := json.Marshal(
		map[string]any{
			"ossRegionId":   regionName,
			"ossBucketName": bucketName,
			"ossObjectName": objectName,
		},
	)
	return c.reviewPicture(string(serviceParameters))
}

func (c *AliClient) ReviewPictureFromBase64(imageBase64 string) (*ReviewImageResponse, error) {
	return nil, errors.New("图片审核接口不支持base64格式")
}

func NewAliClient(conf *SafeReviewConf) (*AliClient, error) {
	_client, err := green20220302.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(conf.Ali.AccessKeyId),
		AccessKeySecret: tea.String(conf.Ali.AccessKeySecret),
		RegionId:        tea.String(conf.Ali.RegionId),
		Endpoint:        tea.String(conf.Ali.Endpoint),
		ConnectTimeout:  tea.Int(3000),
		ReadTimeout:     tea.Int(6000),
	})
	if err != nil {
		return nil, err
	}
	return &AliClient{_client}, nil
}

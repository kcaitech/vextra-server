/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package safereview

type Provider string

const (
	Ali   Provider = "ali"
	Baidu Provider = "baidu"
)

type Client interface {
	ReviewText(text string) (*ReviewTextResponse, error)
	ReviewPictureFromUrl(imageUrl string) (*ReviewImageResponse, error)
	//ReviewPictureFromStorage(regionName string, bucketName string, objectName string) (*ReviewImageResponse, error)
	ReviewPictureFromBase64(imageBase64 string) (*ReviewImageResponse, error)
}

type SafeReviewConf struct {
	Provider Provider `yaml:"provider" json:"provider"`
	Ali      struct {
		AccessKeyId     string `yaml:"accessKeyId" json:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret" json:"accessKeySecret"`
		RegionId        string `yaml:"regionId" json:"regionId"`
		Endpoint        string `yaml:"endpoint" json:"endpoint"`
	} `yaml:"ali" json:"ali"`
	Baidu struct {
		AppId     string `yaml:"appId" json:"appId"`
		ApiKey    string `yaml:"apiKey" json:"apiKey"`
		SecretKey string `yaml:"secretKey" json:"secretKey"`
	} `yaml:"baidu" json:"baidu"`

	TmpPngDir string `yaml:"tmp_png_dir,omitempty" json:"tmp_png_dir,omitempty"`
}

type ReviewTextResult string

const (
	ReviewTextResultPass      ReviewTextResult = "pass"
	ReviewTextResultBlock     ReviewTextResult = "block"
	ReviewTextResultSuspected ReviewTextResult = "suspected"
)

type ReviewTextResponse struct {
	Status ReviewTextResult
	Reason string
	Labels []string
	Words  []string
}

type ReviewImageResult string

const (
	ReviewImageResultPass      ReviewImageResult = "pass"
	ReviewImageResultBlock     ReviewImageResult = "block"
	ReviewImageResultSuspected ReviewImageResult = "suspected"
)

type ReviewImageResultItem struct {
	Reason     string
	Confidence float64
}

type ReviewImageResponse struct {
	Status ReviewImageResult
	Result []ReviewImageResultItem
	Reason string
}

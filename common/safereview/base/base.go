package base

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

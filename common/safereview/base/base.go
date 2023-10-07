package base

type Provider string

const (
	Ali Provider = "ali"
)

type Client interface {
	ReviewText(text string) (*ReviewTextResponse, error)
	ReviewPicture(imageUrl string) (*ReviewPictureResponse, error)
	ReviewPictureFromStorage(regionName string, bucketName string, objectName string) (*ReviewPictureResponse, error)
}

type ReviewTextResult string

const (
	ReviewTextResultPass  ReviewTextResult = "pass"
	ReviewTextResultBlock ReviewTextResult = "block"
)

type ReviewTextResponse struct {
	Status ReviewTextResult
	Reason string
	Labels []string
}

type ReviewPictureResult string

const (
	ReviewPictureResultPass  ReviewPictureResult = "pass"
	ReviewPictureResultBlock ReviewPictureResult = "block"
)

type ReviewPictureResultItem struct {
	Labels     string
	Confidence float64
}

type ReviewPictureResponse struct {
	Status ReviewPictureResult
	Result []ReviewPictureResultItem
}

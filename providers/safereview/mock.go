package safereview

type MockReviewClient struct {
}

func NewMockReviewClient() (*MockReviewClient, error) {
	return &MockReviewClient{}, nil
}

func (c *MockReviewClient) ReviewText(text string) (*ReviewTextResponse, error) {
	return &ReviewTextResponse{Status: ReviewTextResultPass}, nil
}
func (c *MockReviewClient) ReviewPictureFromUrl(imageUrl string) (*ReviewImageResponse, error) {
	return &ReviewImageResponse{Status: ReviewImageResultPass}, nil
}

func (c *MockReviewClient) ReviewPictureFromBase64(imageBase64 string) (*ReviewImageResponse, error) {
	return &ReviewImageResponse{Status: ReviewImageResultPass}, nil
}

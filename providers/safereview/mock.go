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

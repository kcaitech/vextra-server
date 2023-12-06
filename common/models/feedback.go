package models

type FeedbackType uint8 // 反馈类型

const (
	FeedbackTypeReport1 FeedbackType = iota // 举报-欺诈
	FeedbackTypeReport2                     // 举报-色情低俗
	FeedbackTypeReport3                     // 举报-不正当言论
	FeedbackTypeReport4                     // 举报-其他
	FeedbackTypeLast                        // 最后一个
)

// Feedback 反馈
type Feedback struct {
	BaseModel
	UserId        int64        `gorm:"not null" json:"user_id"` // 用户ID
	Type          FeedbackType `gorm:"not null" json:"type"`    // 类型
	Content       string       `json:"content"`                 // 内容
	ImagePathList string       `json:"image_path_list"`         // 图片路径列表，字符串数组json格式
	PageUrl       string       `json:"page_url"`                // 页面URL
}

func (model Feedback) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

package common

import (
	"errors"

	"kcaitech.com/kcserver/services"
)

// GetDocumentBasicInfoById 通过文档ID获取文档的基本信息
func GetDocumentBasicInfoById(documentId string) (*DocumentInfo, error) {
	// 获取数据库连接
	db := services.GetDBModule().DB

	// 创建返回结果结构体
	docInfo := &DocumentInfo{}

	// 执行查询
	query := `
        SELECT d.id as document_id, d.path, d.version_id, dv.last_cmd_ver_id as last_cmd_id
        FROM document d
        INNER JOIN document_version dv ON dv.document_id = d.id AND dv.version_id = d.version_id AND dv.deleted_at IS NULL
        WHERE d.id = ? AND d.deleted_at IS NULL
        LIMIT 1
    `

	// 执行查询并将结果映射到结构体
	err := db.Raw(query, documentId).Scan(docInfo).Error
	if err != nil {
		return nil, err
	}

	// 检查查询结果是否有效
	if docInfo.DocumentId == "" {
		return nil, errors.New("文档不存在")
	}

	return docInfo, nil
}

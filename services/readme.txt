------------------------------------------------------------------------------------------------------------------------
关于Join查询
------------------------------------------------------------------------------------------------------------------------
    tag标签说明：
    --------------------------------------------------------------------------------------------------------------------
    gorm:"embedded;embeddedPrefix:${table}__"
    参数说明：
    table: 表名，若在table标签或join标签中设置了表别名，则此处也需要设为同样的表别名
    --------------------------------------------------------------------------------------------------------------------
    join:"${table};${joinType};${joinFields}"
    参数说明：
    table: 要连接的表名
        格式：
        table_name: 直接设置表名，无别名
        table_name,table_alias: 设置表别名
    joinType: 连接类型
        取值：
        inner: 内连接
        left: 左连接
        right: 右连接
    joinFields: 连接字段
        格式：
        field:
            field = main_table_name.field
        field1,(table_name.)field2:
            field1 = (table_name.)field2
                table_name为可选参数，支持模板参数
                field2支持模板参数和动态参数
        field1,[(table_name.)field2_1 (table_name.)field2_2 (table_name.)field2_n ...]:
            field1 = (table_name.)field2_x
                table_name为可选参数，支持模板参数
                field2_x支持模板参数和动态参数
                会从[(table_name.)field2_1 (table_name.)field2_2 (table_name.)field2_n ...]中选出第一个可用的参数来使用
    --------------------------------------------------------------------------------------------------------------------
    模板参数和动态参数说明:
    模板参数: 以#开头的参数，如#param，最终会被替换为paramArgs["#param"]
        若paramArgs["#param"]为非"#"开头的字符串：
            joinFields: field1,#field2 -> field1 = main_table_name.${paramArgs["#field2"]}
        若paramArgs["#param"]为"#"开头的字符串，则去掉"#"后以如下方式拼接：
            joinFields: field1,#field2 -> field1 ${paramArgs["#field2"][1:]}
            field1和${paramArgs["#field2"]}之间不会有"="连接符，示例：
                joinFields: deleted_at,#is null -> deleted_at is null
        若paramArgs["#param"]为"##"开头的字符串，则去掉"##"后以如下方式拼接：
            joinFields: field1,##var -> field1 var
    动态参数: 以?开头的参数，如?param，最终会被替换为paramArgs["?param"]。对比模板参数，动态最终生成的sql语句中不会与main_table_name关联，而是替换为一个确定的值，类似where语句中的参数
        joinFields: field1,?field2 -> field1 = ${paramArgs["?field2"]}
    --------------------------------------------------------------------------------------------------------------------
    示例：
    type DocumentQueryResItem struct {
        Document Document         `gorm:"embedded;embeddedPrefix:document__" json:"document" join:";inner;id,[#document_id document_id]"`
        User     User             `gorm:"embedded;embeddedPrefix:user__" json:"user" join:";inner;id,[#user_id document.user_id]"`
        Team     *DocumentTeam    `gorm:"embedded;embeddedPrefix:team__" json:"team" join:";left;id,document.team_id"`
        Project  *DocumentProject `gorm:"embedded;embeddedPrefix:project__" json:"project" join:";left;id,document.project_id"`
        DocumentFavorites    DocumentFavorites    `gorm:"embedded;embeddedPrefix:document_favorites__" json:"document_favorites" join:";left;document_id,document.id;user_id,?user_id"`
        DocumentAccessRecord DocumentAccessRecord `gorm:"embedded;embeddedPrefix:document_access_record__" json:"document_access_record" join:";left;user_id,?user_id;document_id,document.id"`
    }

    func (s *DocumentService) FindDocumentByUserId(userId int64) *[]AccessRecordAndFavoritesQueryResItem {
        var result []AccessRecordAndFavoritesQueryResItem
        _ = s.Find(
            &result,
            &ParamArgs{"?user_id": userId},
            &WhereArgs{"document.user_id = ? and (document.project_id is null or document.project_id = 0)", []any{userId}},
            &OrderLimitArgs{"document_access_record.last_access_time desc", 0},
        )
        return &result
    }
------------------------------------------------------------------------------------------------------------------------

package init

import (
	. "protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/snowflake"
)

func Init(config *BaseConfiguration) {
	jwt.Init("")
	snowflake.Init("")
	models.Init(config)
}

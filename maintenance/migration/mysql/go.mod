module mysql

go 1.20

require (
	gorm.io/gorm v1.25.0
	protodesign.cn/kcserver/utils v0.0.0-00010101000000-000000000000
)

require github.com/go-sql-driver/mysql v1.7.0 // indirect

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.5.0
)

replace protodesign.cn/kcserver/utils => ./../../../utils

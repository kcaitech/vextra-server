package snowflake

import (
	"log"
	"kcaitech.com/kcserver/common/snowflake/config"
	s "kcaitech.com/kcserver/utils/snowflake"
)

var snowFlake *s.SnowFlake

func Init(configFilePath string) {
	if snowFlake == nil {
		var err error
		conf := config.LoadConfig(configFilePath)
		if snowFlake, err = s.NewSnowFlake(conf.Snowflake.WorkerId); err != nil {
			log.Fatalln(err)
		}
	}
}

// NextId 获取下一个ID
func NextId() int64 {
	return snowFlake.NextId()
}

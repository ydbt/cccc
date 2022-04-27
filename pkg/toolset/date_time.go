package toolset

import (
	"time"
)

var cstZone *time.Location

func init() {
	cstZone = time.FixedZone("CST", 8*3600) // 东八区
}

func Now() string {
	return time.Now().In(cstZone).Format("2006-01-02 15:04:05")
}

/**
// 它依赖于 IANA Time Zone Database 数据库
var cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海
fmt.Println("SH : ", time.Now().In(cstSh).Format("2006-01-02 15:04:05"))
*/

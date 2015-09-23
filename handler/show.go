package handler

import (
	"log"

	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/setting"
	"github.com/garyburd/redigo/redis"
)

func GetLog(ctx *macaron.Context) {

	logId := ctx.Req.FormValue("logid")

	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}

	result := ""

	str, err := redis.String(c.Do("LPOP", "buildLog:"+logId))
	for i := 0; str != "" || i < 1; i++ {
		result += str
		str, err = redis.String(c.Do("LPOP", "buildLog:"+logId))
	}
	if result == "bye" {
		result = "---end---"
	}
	ctx.Resp.Write([]byte(result))
}

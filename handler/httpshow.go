package handler

import (
	"log"

	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/setting"
	"github.com/garyburd/redigo/redis"
)

func HTTPBuildLog(ctx *macaron.Context) {

	logId := ctx.Req.FormValue("logid")
	count := ctx.QueryInt64("count")

	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	var str []uint8
	strs, err := c.Do("LRANGE", "buildLog:"+logId, count, count+1)
	if err != nil {
		log.Println("[error when get log]", err)
		str = []uint8("error in server")
	}

	if len(strs.([]interface{})) > 0 {
		str = strs.([]interface{})[0].([]uint8)
	} else {
		str = []uint8("")
	}

	ctx.Resp.Write(str)
}

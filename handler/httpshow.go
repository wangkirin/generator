package handler

import (
	"log"

	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/models"
)

func HTTPBuildLog(ctx *macaron.Context) {

	logId := ctx.Req.FormValue("logid")
	count := ctx.QueryInt64("count")

	var str []uint8
	strs, err := models.GetMsgFromList("buildLog:"+logId, count, count+1)
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

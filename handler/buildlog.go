package handler

import (
	"github.com/Unknwon/macaron"
)

func Log(ctx *macaron.Context) {
	if ctx.Params(":protocol") == "http" {
		httpBuildLog(ctx)
	} else if ctx.Params(":protocol") == "ws" {
		websocketBuildLog(ctx)
	} else {
	}
}

package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
)

const (
	BUILDLOGSTR = "BuildLog_%s"
	PUSHLOGSTR  = "PushLog_%s"
)

func Log(ctx *macaron.Context) {
	if ctx.Params(":protocol") == "http" {
		httpBuildLog(ctx)
	} else {
	}
}

func httpBuildLog(ctx *macaron.Context) {

	logId := ctx.Params(":id")
	logType := ctx.Query("type")

	var str []uint8
	var strs []string
	var err error

	if logType == "buildlog" {
		log.Println("in...", fmt.Sprintf(BUILDLOGSTR, logId))
		strs, err = models.GetLogInfo(fmt.Sprintf(BUILDLOGSTR, logId), 0, -1)

		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
			str = []uint8("error in server")
		}
	} else if logType == "pushlog" {
		log.Println("in ...>>>", fmt.Sprintf(PUSHLOGSTR, logId))
		strs, err = models.GetLogInfo(fmt.Sprintf(PUSHLOGSTR, logId), 0, -1)

		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
			str = []uint8("error in server")
		}
	}
	if len(strs) > 0 {
		for _, v := range strs {
			str = append(str, []uint8(v)...)
		}
	} else {
		str = []uint8("")
	}
	ctx.Resp.Write(str)
}

type ReqInfo struct {
	Id string `json:"Id"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var endChan = make(chan bool, 1)

func websocketBuildLog(ctx *macaron.Context) {

	req := ctx.Req.Request
	resp := ctx.Resp
	if req.Method != "GET" {
		http.Error(resp, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	PushLog(ws, ctx.Params(":id"))
}

func PushLog(ws *websocket.Conn, id string) {
}

func PushMsg(ws *websocket.Conn, WSWriter chan []uint8) {
}

func getAllOldLogById(id string, WSWriter chan []uint8) {
}

func startSubscribe(id string, WSWriter chan []uint8) {
}

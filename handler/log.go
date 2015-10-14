package handler

import (
	"log"

	"github.com/Unknwon/macaron"

	"github.com/containerops/generator/models"
)

func Log(ctx *macaron.Context) {
	if ctx.Params(":protocol") == "http" {
		httpBuildLog(ctx)
	} else if ctx.Params(":protocol") == "ws" {
		websocketBuildLog(ctx)
	} else {
	}
}

func httpBuildLog(ctx *macaron.Context) {

	logId := ctx.Params("id")
	count := ctx.QueryInt64("count")

	var str []uint8
	strs, err := models.GetMsgFromList("buildLog:"+logId, count, count+1)
	if err != nil {
		log.Println("[error when get log]", err)
		str = []uint8("error in server")
	}

	if len(strs) > 0 {
		str = []uint8(strs[0])
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

	defer ws.Close()

	var WSWriter = make(chan []uint8, 1024)
	// push data in channel to socket
	go PushMsg(ws, WSWriter)
	// get build log history and push to channel
	getAllOldLogById(id, WSWriter)
	// get new build log and push to channel
	go startSubscribe(id, WSWriter)

	<-endChan
}

func PushMsg(ws *websocket.Conn, WSWriter chan []uint8) {

	defer ws.Close()

	for {
		msg := <-WSWriter

		if msg == nil {
			continue
		}

		// write message to client
		if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Can't send", err.Error())
			break
		}

		if string(msg) == "bye" {
			break
		}
	}

	endChan <- true

}

func getAllOldLogById(id string, WSWriter chan []uint8) {
	strs, err := models.GetMsgFromList("buildLog:"+id, int64(0), int64(-1))
	if err != nil {
		log.Println("[error when get history log]", err)
	}

	for _, str := range strs {
		WSWriter <- []uint8(str)
	}
}

func startSubscribe(id string, WSWriter chan []uint8) {
	msgChan := models.SubscribeChannel("buildLog:" + id)
	for {
		msg := <-msgChan
		WSWriter <- []uint8(msg)
		if "bye" == msg {
			break
		}
	}

}

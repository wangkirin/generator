package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/containerops/generator/models"
	"github.com/gorilla/websocket"
)

type ReqInfo struct {
	Id string `json:"Id"`
}

func WSbuildLog(ctx *macaron.Context) {

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

	ReceiveId(ws)
}

func ReceiveId(ws *websocket.Conn) {

	defer ws.Close()

	for {
		_, message, err := ws.ReadMessage()

		if err != nil {
			log.Println("Can't receive %s", err.Error())
			break
		}

		if string(message) == "" {
			log.Println("Receive message is null")
			break
		}

		var info ReqInfo
		if err := json.Unmarshal(message, &info); err != nil {
			log.Println(err.Error())
		} else {
			var WSWriter = make(chan []uint8, 1024)
			// push data in channel to socket
			go PushMsg(ws, WSWriter)
			// get build log history and push to channel
			getAllOldLogById(info.Id, WSWriter)
			// get new build log and push to channel
			go startSubscribe(info.Id, WSWriter)
		}
	}
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
	}

}

func getAllOldLogById(id string, WSWriter chan []uint8) {
	strs, err := models.GetMsgFromList("buildLog:"+id, int64(0), int64(-1))
	if err != nil {
		log.Println("[error when get history log]", err)
	}

	for _, str := range strs.([]interface{}) {
		WSWriter <- str.([]uint8)
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

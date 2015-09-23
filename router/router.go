package router

import (
	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {

	m.Get("/", handler.IndexHandler)
	m.Get("/ws", handler.ServeWs)
	m.Get("/ws_getbuildlog", handler.WSGetLog)
	m.Post("/show", handler.GetLog)
	m.Post("/buildreq", handler.SendBuildReq)
}

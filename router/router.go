package router

import (
	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {

	m.Get("/", handler.IndexHandler)
	m.Get("/ws", handler.ServeWs)
	m.Get("/show", handler.GetLog)
}

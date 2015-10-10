package router

import (
	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {

	//m.Get("/", handler.IndexHandler)
	//m.Get("/ws", handler.WSServer)
	m.Get("/wsbuildlog", handler.WSbuildLog)
	m.Post("/httpbuildlog", handler.HTTPBuildLog)
	m.Post("/httpbuild", handler.HTTPBuild)
}

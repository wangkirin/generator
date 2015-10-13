package router

import (
	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {
	m.Group("/v1", func() {
		m.Group("/build", func() {
			m.Get("/", handler.Build)
			m.Get("/log/:protocol/:id", handler.Log)
		})
	})
}

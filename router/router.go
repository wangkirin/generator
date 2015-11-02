package router

import (
	"github.com/containerops/generator/handler"
	"gopkg.in/macaron.v1"
)

func SetRouters(m *macaron.Macaron) {
	m.Group("/b1", func() {
		m.Group("/build", func() {
			m.Get("/", handler.Build)
			m.Get("/log/:protocol/:id", handler.Log)
		})
	})
}

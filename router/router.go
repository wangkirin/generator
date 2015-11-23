package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {
	m.Group("/b1", func() {
		m.Group("/build", func() {
			m.Post("/", handler.Build)

			m.Post("/deamon", handler.AddDeamon)
			m.Delete("/deamon", handler.DelDeamon)
		})

		m.Group("/log", func() {
			m.Get("/:protocol/:id", handler.log)
		})

		m.Group("/job", func() {
			m.Get("/all", handler.GetAllJobs)
		})

		m.Group("/daemon", func() {
			m.Get("/all", handler.GetAllDeamons)
			m.Get("/info/:id", handler.GetInfo)
		})
	})
}

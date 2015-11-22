package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/handler"
)

func SetRouters(m *macaron.Macaron) {
	m.Group("/b1", func() {
		m.Group("/build", func() {
			m.Post("/", handler.Build)
			m.Post("/adddeamon", handler.AddDeamon)

			m.Delete("/deldeamon", handler.DelDeamon)

			m.Get("/log/:protocol/:id", handler.Log)
			m.Get("/info", handler.GetInfo)
			m.Get("/getalljobs", handler.GetAllJobs)
			m.Get("/getalldeamon", handler.GetAllDeamons)
		})
	})
}

package middleware

import (
	"gopkg.in/macaron.v1"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Map(Log)
	m.Use(logger())
	m.Use(macaron.Recovery())
}

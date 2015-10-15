package middleware

import (
	"github.com/Unknwon/macaron"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Map(Log)
	m.Use(logger())
	m.Use(macaron.Recovery())
}

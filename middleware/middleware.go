package middleware

import (
	"github.com/Unknwon/macaron"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Recovery())

	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger())
}

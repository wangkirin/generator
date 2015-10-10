package middleware

import (
	"github.com/Unknwon/macaron"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	m.Use(macaron.Recovery())

	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger())

}

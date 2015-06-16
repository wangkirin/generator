package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/generator/middleware"
	"github.com/containerops/generator/router"
)

func SetGeneratorMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
}

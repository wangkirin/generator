package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/generator/middleware"
	"github.com/containerops/generator/router"
	"github.com/containerops/wrench/setting"
)

func SetGeneratorMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
	if setting.RunMode == "dev" {
		m.Use(macaron.Static("tests"))
	}
}

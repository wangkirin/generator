package web

import (

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/middleware"
	"github.com/containerops/generator/router"
	"github.com/containerops/wrench/setting"
)

func SetGeneratorMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
	//static
	if setting.RunMode == "dev" {
		m.Use(macaron.Static("external"))
	}

}

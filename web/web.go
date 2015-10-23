package web

import (
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/handler"
	"github.com/containerops/generator/middleware"
	"github.com/containerops/generator/modules/build"
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

	if err := modules.LoadBuildList("/conf/pool.json"); err != nil {
		fmt.Printf("Read build pool config file /conf/pool.json error: %s", err.Error())
	}

	handler.InitHandlerList()

}

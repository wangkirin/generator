package middleware

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Map(Log)
	m.Use(logger())
	m.Use(macaron.Recovery())
	db.InitDB(setting.DBURI, setting.DBPasswd, setting.DBDB)
	models.InitJob()
	models.InitDeamon()
}

package handler

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
)

func GetInfo(ctx *macaron.Context) string {
	host := ctx.Query("host")
	port := ctx.Query("port")

	deamon := new(models.Deamon)
	return deamon.Info(host, port)
}

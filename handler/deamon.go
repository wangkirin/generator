package handler

import (
	"encoding/json"
	"log"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
)

func AddDeamon(ctx *macaron.Context) string {
	bytes, err := ctx.Req.Body().Bytes()

	deamonInfo := new(models.Deamon)
	err = json.Unmarshal(bytes, deamonInfo)
	if err != nil {
		log.Println("[errorInfo]:error when unmarshal add deamon request body:", err.Error())
		return "failed"
	}

	err = deamonInfo.Add(deamonInfo.Host, deamonInfo.Port)
	if err != nil {
		log.Println("[errorInfo]:error when add deamon:", err.Error())
		return "failed"
	}
	return "success"
}

func DelDeamon(ctx *macaron.Context) string {
	bytes, err := ctx.Req.Body().Bytes()

	deamonInfo := new(models.Deamon)
	err = json.Unmarshal(bytes, deamonInfo)
	if err != nil {
		log.Println("[errorInfo]:error when unmarshal add deamon request body:", err.Error())
		return "failed"
	}

	err = deamonInfo.Del(deamonInfo.Host, deamonInfo.Port)
	if err != nil {
		log.Println("[errorInfo]:error when add deamon:", err.Error())
		return "failed"
	}
	return "success"

}

func GetAllDeamons() string {
	strs, err := models.GetAllDeamons()
	result := ""
	if err != nil {
		log.Println("[errorInfo]:error when get all deamon:" + err.Error())
		return ""
	}
	for k, v := range strs {
		if k%2 == 1 {
			result += v + "\n"
		}
	}
	return result
}

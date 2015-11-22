package handler

import (
	"log"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
)

func GetAllJobs(ctx *macaron.Context) string {
	jobs, err := models.GetAllJobs()
	if err != nil {
		log.Println("[errorInfo]:error when get jobs:", err.Error())
		return ""
	}
	result := ""
	for _, v := range jobs {
		result += v + "\n"
	}
	return result
}

package handler

import (
	"encoding/json"
	"log"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
	. "github.com/containerops/generator/modules/build"
)

const (
	DEFAULT_DockerfileName = ""
	DEFAULT_RemoteURL      = ""
	DEFAULT_RepoName       = ""
	DEFAULT_SuppressOutput = true
	DEFAULT_NoCache        = true
	DEFAULT_Remove         = true
	DEFAULT_ForceRemove    = true
	DEFAULT_Pull           = true
)

type RequestStruct struct {
	Repo           string `json:"repo"`
	Tag            string `json:"tag"`
	DockerFile     string `json:"dockerfile"`
	RemoteURL      string `json:"remote_url"`
	RepoName       string `json:"repo_name"`
	SuppressOutput bool   `json:"suppress_output"`
	NoCache        bool   `json:"no_cache"`
	Remove         bool   `json:"remove"`
	ForceRemove    bool   `json:"force_remove"`
	Pull           bool   `json:"pull"`
}

func Build(ctx *macaron.Context) string {

	job := new(models.Job)
	buildImgConfig := new(BuildImage)
	comingInfo := new(RequestStruct)
	bodyByte, err := ctx.Req.Body().Bytes()
	bodyStr := string(bodyByte)
	if err != nil {
		log.Println("[errorInfo]:error when get request body:", err.Error())
		return ""
	}

	err = json.Unmarshal(bodyByte, comingInfo)
	if err != nil {
		log.Println("[errorInfo]:error when unmarshal request body:", err.Error())
		return ""
	}

	err = json.Unmarshal(bodyByte, buildImgConfig)
	if err != nil {
		log.Println("[errorInfo]:error when unmarshal request body:", err.Error())
		return ""
	}

	if !strings.Contains(bodyStr, "suppress_output") {
		buildImgConfig.SuppressOutput = true
	}
	if !strings.Contains(bodyStr, "no_cache") {
		buildImgConfig.NoCache = true
	}
	if !strings.Contains(bodyStr, "remove") {
		buildImgConfig.Remove = true
	}
	if !strings.Contains(bodyStr, "force_remove") {
		buildImgConfig.ForceRemove = true
	}
	if !strings.Contains(bodyStr, "pull") {
		buildImgConfig.Pull = true
	}

	config, err := json.Marshal(buildImgConfig)
	if err != nil {
		log.Println("[errorInfo]:error when befroe add job:", err.Error())
		return ""
	}

	str, err := job.Add(comingInfo.Repo, comingInfo.Tag, comingInfo.DockerFile, string(config))
	if err != nil {
		log.Panicln("[errorInfo]:error when add job:", err.Error())
		return ""
	}
	return str
}

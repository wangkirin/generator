package modules

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/containerops/generator/models"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

func LoadBuildList(path string) error {
	er := db.InitDB(setting.DBURI, setting.DBPasswd, setting.DBDB)
	if er != nil {
		log.Println(er)
	}
	result := readConfigFile(path)
	var list BuilderList
	if err := json.Unmarshal([]byte(result), &list); err != nil {
		return err
	} else {
		saveToRedis(list)
	}
	return nil
}

func readConfigFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	result, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return string(result)
}

type BuilderList struct {
	Dockers []BuilderInfo `json:"docker"`
}

type BuilderInfo struct {
	IP   string `json:"url"`
	PORT string `json:"port"`
}

func saveToRedis(list BuilderList) {
	for _, v := range list.Dockers {
		err := models.SaveMsgToSet("DockerList", v.IP+":"+v.PORT)
		if err != nil {
			log.Println("err when save docker build list", err)
		}
	}

}

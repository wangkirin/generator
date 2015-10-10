package modules

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/containerops/generator/models"
	"github.com/containerops/wrench/setting"
	"github.com/garyburd/redigo/redis"
)

func LoadBuildList(path string) error {
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
	Rkt     []BuilderInfo `json:"rkt"`
}

type BuilderInfo struct {
	IP   string `json:"url"`
	PORT string `json:"port"`
}

func saveToRedis(list BuilderList) {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	for _, v := range list.Dockers {
		err = models.SaveMsgToSet("DockerList", v.IP+":"+v.PORT)
		if err != nil {
			log.Println("err when save docker build list", err)
		}
	}

	for _, v := range list.Rkt {
		err = models.SaveMsgToSet("RKTList", v.IP+":"+v.PORT)
		if err != nil {
			log.Println("err when save docker build list", err)
		}

	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/containerops/generator/cmd"
	"github.com/containerops/wrench/setting"
	"github.com/garyburd/redigo/redis"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := setting.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read config error: %v", err.Error())
	}

	result := readConfigFile("./conf/pool.json")
	var list BuilderList
	if err := json.Unmarshal([]byte(result), &list); err != nil {
		log.Fatalln(err)
	} else {
		saveToRedis(list)
	}

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
		_, err = c.Do("SADD", "DockerList", v.IP+":"+v.PORT)
		if err != nil {
			log.Println("err when save docker build list", err)
		}
	}

	for _, v := range list.Rkt {
		_, err = c.Do("SADD", "RKTList", v.IP+":"+v.PORT)
		if err != nil {
			log.Println("err when save docker build list", err)
		}

	}
}

func main() {
	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}

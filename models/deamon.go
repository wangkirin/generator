package models

import (
	"errors"
	"io"
	"log"
	"strings"

	. "github.com/containerops/generator/modules/build"
	"github.com/containerops/wrench/db"
)

const (
	GEN_DEAMON          = "GEN_DEAMON"
	GEN_DEAMON_NEED_DEL = "GEN_DEAMON_NEED_DEL"
)

type Deamon struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

var (
	DeamonChans = make(chan *Deamon, 65535)
)

func InitDeamon() {
	deamons, err := db.Client.HGetAll(GEN_DEAMON).Result()
	if err != nil {
		panic(err)
	}

	for _, value := range deamons {
		hpa := strings.Split(value, ":")

		if len(hpa) != 2 {
			panic(errors.New("can't init deamons by db!"))

		}

		deamon := &Deamon{
			Host: hpa[0],
			Port: hpa[1],
		}
		DeamonChans <- deamon
	}
}

func (d *Deamon) Add(host, port string) (err error) {

	_, err = db.Client.HSet(GEN_DEAMON, host+":"+port, host+":"+port).Result()
	if err != nil {
		return err
	}

	deamon := &Deamon{
		Host: host,
		Port: port,
	}
	DeamonChans <- deamon

	return nil
}

func (d *Deamon) Del(host, port string) (err error) {

	_, err = db.Client.HSet(GEN_DEAMON_NEED_DEL, host+":"+port, host+":"+port).Result()
	if err != nil {
		return err
	}
	return nil
}

func (d *Deamon) Info(host, port string) (info string) {
	client, _ := NewDockerClient(host+":"+port, nil)
	infoReader, err := client.Info()
	if err != nil {
		log.Println("[errorInfo]:error when get docker info:" + err.Error())
		return ""
	}
	buf := make([]byte, 4096)
	str := ""
	for {
		n, err := infoReader.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("[errorInfo]:error when read docker info:" + err.Error())
			return ""
		}

		if n == 0 {
			return str
		}

		str += string(buf[:n])
	}
	return ""
}

func GetAllDeamons() ([]string, error) {
	return db.Client.HGetAll(GEN_DEAMON).Result()
}

func (d *Deamon) restore(host, port string) (err error) {

	exists, err := db.Client.HExists(GEN_DEAMON_NEED_DEL, host+":"+port).Result()
	if err != nil {
		return err
	}

	if exists {
		db.Client.HDel(GEN_DEAMON, host+":"+port)
		db.Client.HDel(GEN_DEAMON_NEED_DEL, host+":"+port)
		return nil
	}

	deamon := &Deamon{
		Host: host,
		Port: port,
	}
	DeamonChans <- deamon
	return nil
}

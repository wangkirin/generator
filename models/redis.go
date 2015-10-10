package models

import (
	"log"

	"github.com/containerops/wrench/setting"
	"github.com/garyburd/redigo/redis"
)

func GetRandomOneFromSet(setName string) (string, error) {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	return redis.String(c.Do("SRANDMEMBER", setName))
}

func GetMsgFromList(listName string, start, end int64) (interface{}, error) {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	return c.Do("LRANGE", listName, start, end)
}

func SaveMsgToSet(setName, msg string) error {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	_, err = c.Do("SADD", setName, msg)
	return err
}

func PushMsgToList(listName, msg string) error {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	_, err = c.Do("RPUSH", listName, msg)
	return err
}

func SubscribeChannel(channelName string) chan string {
	msgChan := make(chan string, 30)
	go receiveMsgFromChannel(channelName, msgChan)
	return msgChan
}

func receiveMsgFromChannel(channelName string, msgChan chan string) {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	psc := redis.PubSubConn{c}
	psc.Subscribe(channelName)
	var isEnd = false
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			msgChan <- string(v.Data)
			if testSliceEq(v.Data, []byte("bye")) {
				isEnd = true
			}
		}
		if isEnd {
			break
		}
	}
}

func testSliceEq(a, b []uint8) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func PublishMsg(channelName, msg string) error {
	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	_, err = c.Do("PUBLISH", channelName, msg)
	return err
}

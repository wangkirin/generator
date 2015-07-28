package docker

import (
	"errors"
	"github.com/containerops/generator/setting"
	"math/rand"
	"time"
)

func GetOneDaemon() (map[string]string, error) {
	if len(setting.GenPool.Docker) <= 0 {
		return nil, errors.New("Pool is Empty！！！")
	}
	oneDaemon := setting.GenPool.Docker[generateRandomNumber(0, len(setting.GenPool.Docker))]
	return oneDaemon, nil
}

//生成1个[start,end)结束的随机数
func generateRandomNumber(start int, end int) int {
	//范围检查
	if end < start {
		return -1
	}

	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Intn((end - start)) + start
}

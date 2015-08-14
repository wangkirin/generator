package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	generatorUtils "github.com/containerops/wrench/utils"
)

var (
	buildInfoChan chan io.Reader
)

func Test_GetOnDaemon(t *testing.T) {
	Convey("Test Docker Deamon", t, func() {

		PoolInfo, err := GetOneDaemon()
		fmt.Println("PoolInfo\n")
		fmt.Println(PoolInfo)
		So(PoolInfo, ShouldNotBeNil)
		So(err, ShouldBeNil)

		daemonUrl := PoolInfo["url"]
		daemonPort := PoolInfo["port"]
		fmt.Println(daemonUrl, ":", daemonPort)

		//connection  docker  deamon
		endpoint := fmt.Sprint("tcp://", daemonUrl, ":", daemonPort)
		client, _ := generatorUtils.NewClient(endpoint)
		imgs, _ := client.ListImages(generatorUtils.ListImagesOptions{All: false})

		// Print daemon images list
		for _, img := range imgs {
			fmt.Println("ID: ", img.ID)
			fmt.Println("RepoTags: ", img.RepoTags)
			fmt.Println("Created: ", img.Created)
			fmt.Println("Size: ", img.Size)
			fmt.Println("VirtualSize: ", img.VirtualSize)
			fmt.Println("ParentId: ", img.ParentID)
		}
		fmt.Println("-------------------------------------------------------------")
		// set docker build opt
		t := time.Now()
		inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
		tr := tar.NewWriter(inputbuf)
		content, err := ioutil.ReadFile(fmt.Sprint("Dockerfile"))

		tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(content)), ModTime: t, AccessTime: t, ChangeTime: t})
		tr.Write(content)
		tr.Close()
		opts := generatorUtils.BuildImageOptions{
			Name:         "test2",
			InputStream:  inputbuf,
			OutputStream: outputbuf,
		}
		//run deamon remot build
		client.BuildImage(opts)
		fmt.Println(outputbuf.String())
		/*
			go func(opts modules.BuildImageOptions) {
				client.BuildImage(opts)
				fmt.Println(outputbuf.String())
			}(opts)

			p := make([]byte, 5)
			var total int
			for {
				n, err := outputbuf.Read(p)
				if err == io.EOF && total > 0 {
					fmt.Println(total) //5
					break
				}
				fmt.Print(len(p))
				total = total + n
			}*/

		//client.Push

	})

}

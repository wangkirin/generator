package cmd

import (
	"archive/tar"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	. "github.com/containerops/generator/modules"
	"github.com/containerops/generator/setting"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"os"
)

var CmdWebSocket = cli.Command{
	Name:        "websocket",
	Usage:       "start generator websocket service",
	Description: "get Dockerfile,send build image info.",
	Action:      runWebSocket,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "websocket service listen ip, default is 0.0.0.0; if listen with Unix Socket, the value is sock file path.",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 20000,
			Usage: "websocket service listen at port 20000;",
		},
	},
}

//make chan buffer write to websocket
var ws_writer = make(chan string, 100000)

func SendMsg(ws *websocket.Conn) {

	go func() {

		for {
			msg := <-ws_writer

			if err := websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("Can't send", err.Error())
				break
			}
		}
	}()

}

type Image struct {
	Name       string `json:"name"`
	Dockerfile string `json:"dockerfile"`
}

func ReceiveMsg(ws *websocket.Conn) {
	SendMsg(ws)

	log.Println("In ReceiveMsg")
	var msg string

	for {
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			log.Println("Can't receive %s", err.Error())
			return
		}

		log.Println("before json ", msg)

		var image Image
		if err := json.Unmarshal([]byte(msg), &image); err != nil {
			log.Println(err.Error())
		}

		dockerfile, err := base64.StdEncoding.DecodeString(image.Dockerfile)
		if err != nil {

			log.Println("[ErrorInfo]", err.Error())
		}
		log.Println(string(dockerfile))

		//save msg to file
		tempfile := "Dockerfile"
		fout, err := os.Create(tempfile)
		defer fout.Close()
		if err != nil {
			fmt.Println(tempfile, err)
			return
		}
		fout.WriteString(string(dockerfile) + "\r\n")

		//tar file to tar
		fw, err := os.Create(setting.DOCKERFILEPATH)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer fw.Close()

		tw := tar.NewWriter(fw)
		defer tw.Close()

		fo, err := os.Open(tempfile)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fs, err := os.Stat(tempfile)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//tar header
		h := new(tar.Header)
		h.Name = tempfile
		h.Size = fs.Size()
		h.Mode = int64(fs.Mode())
		h.ModTime = fs.ModTime()

		//write tar header
		err = tw.WriteHeader(h)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// copy file content
		_, err = io.Copy(tw, fo)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		log.Println("image.Name", image.Name)

		//build docker
		BuildDockerImage(image.Name)

	}
}

func BuildDockerImage(imagename string) {

	// Init the clientdocker, _ := NewDockerClient(testDockerUrl, nil)
	docker, _ := NewDockerClient(setting.DOCKERURL, nil)
	// Build a docker image
	// some.tar contains the build context (Dockerfile any any files it needs to add/copy)
	dockerBuildContext, err := os.Open(setting.DOCKERFILEPATH)
	defer dockerBuildContext.Close()

	buildImageConfig := &BuildImage{
		Context:        dockerBuildContext,
		RepoName:       imagename,
		SuppressOutput: true,
	}

	reader, err := docker.BuildImage(buildImageConfig)
	if err != nil {
		fmt.Println(err.Error())
	}

	buf := make([]byte, 4096)

	for {

		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			ws_writer <- "bye"
			break
		}

		//ctx.Write(buf[:n])
		ws_writer <- string(buf[:n])
	}

}

func runWebSocket(c *cli.Context) {

	//start websocket service
	http.Handle("/", websocket.Handler(ReceiveMsg))

	http.ListenAndServe(":20000", nil)
}

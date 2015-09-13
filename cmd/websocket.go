package cmd

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/gorilla/websocket"

	. "github.com/containerops/generator/modules"
	"github.com/containerops/generator/setting"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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
var ws_writer = make(chan string, 1024)

func SendMsg(ws *websocket.Conn) {

	defer ws.Close()

	for {
		msg := <-ws_writer

		if msg == "" {
			continue
		}

		// write message to client
		if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Can't send", err.Error())
			break
		}
	}

}

type BuildImageInfo struct {
	Name       string `json:"name"`
	Dockerfile string `json:"dockerfile"`
}

func ReceiveMsg(ws *websocket.Conn) {

	defer ws.Close()

	var msg string

	for {
		_, message, err := ws.ReadMessage()

		if err != nil {
			log.Println("Can't receive %s", err.Error())
			break
		}

		if message == "" {
			log.Println("Receive message is null")
			break
		}

		var buildImageInfo BuildImageInfo
		if err := json.Unmarshal(message, &buildImageInfo); err != nil {
			log.Println(err.Error())
		}

		dockerfileBytes, err := base64.StdEncoding.DecodeString(buildImageInfo.Dockerfile)
		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
		}
		// Create a buffer to write our archive to.
		buf := new(bytes.Buffer)

		// Create a new tar archive.
		tw := tar.NewWriter(buf)

		// Add some files to the archive.
		var files = []struct {
			Name, Body string
		}{
			{"Dockerfile", string(dockerfileBytes)},
		}
		for _, file := range files {
			hdr := &tar.Header{
				Name: file.Name,
				Mode: 0600,
				Size: int64(len(file.Body)),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				log.Fatalln(err)
			}
			if _, err := tw.Write([]byte(file.Body)); err != nil {
				log.Fatalln(err)
			}
		}
		// Make sure to check the error on Close.
		if err := tw.Close(); err != nil {
			log.Fatalln(err)
		}
		tarReader := bytes.NewReader(buf.Bytes())
		//build docker
		BuildDockerImage(buildImageInfo.Name, tarReader)

	}
}

func BuildDockerImage(imageName string, dockerfileTarReader io.Reader) {

	log.Println("setting.DockerGenUrl:::", setting.DockerGenUrl)

	dockerClient, _ := NewDockerClient(setting.DockerGenUrl, nil)

	buildImageConfig := &BuildImage{
		Context:        dockerfileTarReader,
		RepoName:       imageName,
		SuppressOutput: true,
	}

	reader, err := dockerClient.BuildImage(buildImageConfig)
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

		ws_writer <- string(buf[:n])
	}

}

// ServeWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go SendMsg(ws)
	ReceiveMsg(ws)
}

func runWebSocket(c *cli.Context) {
	//start websocket service
	http.HandleFunc("/", ServeWs)
	http.ListenAndServe(":20000", nil)
}

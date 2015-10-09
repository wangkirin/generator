package handler

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/containerops/generator/models"
	. "github.com/containerops/generator/modules"
	"github.com/containerops/wrench/utils"
)

func HTTPBuild(ctx *macaron.Context) {

	dockerfileBytes, err := base64.StdEncoding.DecodeString(ctx.Req.FormValue("dockerfile"))

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

	//geneGuid for buildlog
	tag := geneGuid()

	//build docker
	go BuildDockerImageStartByHTTPReq(ctx.Query("imagename"), tarReader, tag)

	ctx.Write([]byte(tag))
}

func geneGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		log.Println("err when get geneGuid:", err)
	}
	return utils.MD5(base64.URLEncoding.EncodeToString(b))
}

func BuildDockerImageStartByHTTPReq(imageName string, dockerfileTarReader io.Reader, tag string) {

	dockerUrl, err := models.GetRandomOneFromSet("DockerList")
	if err != nil {
		log.Fatalln("err when get docker list", err)
		return
	}

	dockerClient, _ := NewDockerClient(dockerUrl, nil)

	buildImageConfig := &BuildImage{
		Context:        dockerfileTarReader,
		RepoName:       imageName,
		SuppressOutput: true,
	}

	reader, err := dockerClient.BuildImage(buildImageConfig)
	if err != nil {
		log.Println(err.Error())
	}

	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if strings.Contains(string(buf[:n]), `"stream":"Successfully built`) {
			dockerClient.PushImage(buildImageConfig)
		}

		if 0 == n {
			err = models.PushMsgToList("buildLog:"+tag, "bye")
			if err != nil {
				log.Println("err when write to redis:", err)
			}

			err = models.PublishMsg("buildLog:"+tag, "bye")
			if err != nil {
				log.Println("err when publish build log:", err)
			}

			break
		}

		err = models.PushMsgToList("buildLog:"+tag, string(buf[:n]))
		if err != nil {
			log.Println("err when write to redis:", err)
		}

		err = models.PublishMsg("buildLog:"+tag, string(buf[:n]))
		if err != nil {
			log.Println("err when write to redis:", err)
		}

	}
}

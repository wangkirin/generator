package handler

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"

	"github.com/Unknwon/macaron"
	. "github.com/containerops/generator/modules"
	"github.com/containerops/generator/setting"
	"github.com/containerops/wrench/utils"
	"github.com/garyburd/redigo/redis"
)

func SendBuildReq(ctx *macaron.Context) {

	//dockerfileBytes, err := base64.StdEncoding.DecodeString(ctx.Query("dockerfile"))
	dockerfileBytes, err := base64.StdEncoding.DecodeString(ctx.Req.FormValue("dockerfile"))

	log.Println("from broswer============================>\n", ctx.Req.FormValue("dockerfile"))
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

	c, err := redis.Dial("tcp", setting.DBURI, redis.DialPassword(setting.DBPasswd), redis.DialDatabase(int(setting.DBDB)))
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	dockerClient, _ := NewDockerClient(setting.DockerGenUrl, nil)

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
		if 0 == n {
			_, err = c.Do("RPUSH", "buildLog:"+tag, "bye")
			if err != nil {
				log.Println("err when write to redis:", err)
			}
			break
		}
		log.Println("=============>", string(buf[:n]))
		_, err = c.Do("RPUSH", "buildLog:"+tag, string(buf[:n])+"<br/>")
		if err != nil {
			log.Println("err when write to redis:", err)
		}
	}
}

package handler

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
	. "github.com/containerops/generator/modules/build"
	"github.com/containerops/wrench/utils"
)

type Job struct {
	Name       string `json:"name"`
	DockerFile string `json:"dockerfile"`
	Tag        string `json:"tag"`
}

var freeWorkerList chan string
var busyWorkerList []string
var unhandleJobList chan *Job

func Build(ctx *macaron.Context) string {

	job := new(Job)
	job.Name = ctx.Query("imagename")
	job.DockerFile = ctx.Query("dockerfile")
	job.Tag = geneGuid()
	addJob(job)

	return job.Tag
}

func geneGuid() string {
	guidBuff := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, guidBuff); err != nil {
		log.Println("[ErrorInfo]", err.Error())
	}
	return utils.MD5(base64.URLEncoding.EncodeToString(guidBuff))
}

func BuildDockerImageStartByHTTPReq(worker, imageName string, dockerfileTarReader io.Reader, tag string) {

	dockerClient, _ := NewDockerClient(worker, nil)

	buildImageConfig := &BuildImage{
		Context:        dockerfileTarReader,
		RepoName:       imageName,
		SuppressOutput: true,
	}

	reader, err := dockerClient.BuildImage(buildImageConfig)
	if err != nil {
		log.Println("[ErrorInfo]", err.Error())
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
				log.Println("[ErrorInfo]", err.Error())
			}

			err = models.PublishMsg("buildLog:"+tag, "bye")
			if err != nil {
				log.Println("[ErrorInfo]", err.Error())
			}
			finishJob(worker)
			break
		}

		err = models.PushMsgToList("buildLog:"+tag, string(buf[:n]))
		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
		}

		err = models.PublishMsg("buildLog:"+tag, string(buf[:n]))
		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
		}
	}
}

func InitHandlerList() {
	freeWorkerList = make(chan string, 65535)
	unhandleJobList = make(chan *Job, 65535)

	// get all worker list
	workers, err := models.GetMsgFromList("DockerList", int64(0), int64(-1))
	if err != nil {
		log.Println("[ErrorInfo]", err.Error())
	}

	for _, v := range workers {
		addWorker(v)
	}

	// read unhandleJobList from redis
	jobs, err := models.GetMsgFromList("DockerJobList", int64(0), int64(-1))
	if err != nil {
		log.Println("[ErrorInfo]", err.Error())
	}

	for _, v := range jobs {
		temp := new(Job)
		err := json.Unmarshal([]byte(v), temp)
		if err != nil {
			log.Println("[ErrorInfo]", err.Error())
		}
		unhandleJobList <- temp
	}

	go handleJob()
}

// add a docker machine to worker list
func addWorker(addr string) {
	models.PushMsgToList("FreeWorkerList", addr)
	freeWorkerList <- addr
}

// add a job to job list
func addJob(job *Job) {
	msg, err := json.Marshal(job)
	if err != nil {
		log.Println("[ErrorInfo]", err.Error())
	}

	models.PushMsgToList("DockerJobList", string(msg))
	unhandleJobList <- job
}

// handle job when has a free docker machine and an unhandle job
func handleJob() {
	for {
		worker := <-freeWorkerList
		models.MoveFromListByValue("FreeWorkerList", worker, 0)
		job := <-unhandleJobList
		jobStr, err := json.Marshal(job)
		if err != nil {
			log.Println("[ErrorInfo]", err)
		}

		models.MoveFromListByValue("DockerJobList", string(jobStr), 0)
		busyWorkerList = append(busyWorkerList, worker)
		models.PushMsgToList("BusyWorkerList", worker)

		go BuildDockerImageStartByHTTPReq(worker, job.Name, tarDockerFile(job.DockerFile), job.Tag)
	}
}

// add a docker machine to worker list
func finishJob(addr string) {

	models.MoveFromListByValue("BusyWorkerList", addr, 0)

	// add worker to freeWorkerList
	freeWorkerList <- addr

	// cycle busyWorkerList to remove this worker
	temp := make(chan string, 65535)

	for _, v := range busyWorkerList {
		if v != addr {
			temp <- v
		}
	}

	freeWorkerList = temp
}

func tarDockerFile(dockerfile string) io.Reader {

	dockerfileBytes, err := base64.StdEncoding.DecodeString(dockerfile)

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

	return tarReader
}

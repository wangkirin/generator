package handler

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/macaron.v1"

	"github.com/containerops/generator/models"
	. "github.com/containerops/generator/modules/build"
	"github.com/containerops/wrench/utils"
)

type Job struct {
	Name string `json:"name"`
	// Mode :set tar or dockerfile to enable different mode
	Mode        string     `json:mode`
	DockerFile  string     `json:"dockerfile"`
	ImageConfig BuildImage `json:"buildimage"`
	// Context : when mode == "dockerfile", context is the content of dockerfile
	// when mode == "archive", context is the path to the archive path
	Context string `"json:context"`
	Tag     string `json:"tag"`
}

var freeWorkerList chan string
var busyWorkerList []string
var unhandleJobList chan *Job

func Build(ctx *macaron.Context) string {
	job := new(Job)

	job.Mode = ctx.Query("mode")
	job.Name = ctx.Query("imagename")
	job.Context = ctx.Query("context")

	setImage(job, ctx)

	job.Tag = geneGuid()

	addJob(job)

	return job.Tag
}

// SetImage , set build config to job.ImageConfig from restapi of /build
func setImage(job *Job, ctx *macaron.Context) {
	job.ImageConfig.Dockerfile = ctx.Query("dockerfile")
	job.ImageConfig.RemoteURL = ctx.Query("remoteurl")
	job.ImageConfig.RepoName = ctx.Query("imagename")
	job.ImageConfig.SuppressOutput = ctx.Query("suppressoutput")
	job.ImageConfig.NoCache = ctx.Query("nocache")
	job.ImageConfig.ForceRemove = ctx.Query("forceremove")
	job.ImageConfig.Pull = ctx.Query("pull")
	job.ImageConfig.Memory = ctx.Query("memory")
	job.ImageConfig.MemorySwap = ctx.Query("memoryswap")
	job.ImageConfig.CpuShares = ctx.Query("cpushares")
	job.ImageConfig.CpuPeriod = ctx.Query("cpuperiod")
	job.ImageConfig.CpuQuota = ctx.Query("cpuquota")
	job.ImageConfig.CpuSetCpus = ctx.Query("cpusetcpus")
	job.ImageConfig.CpuSetMems = ctx.Query("cpusetmems")
	job.ImageConfig.CgroupParent = ctx.Query("cgroupparent")
}

func geneGuid() string {
	guidBuff := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, guidBuff); err != nil {
		log.Println("[ErrorInfo]", err.Error())
	}
	return utils.MD5(base64.URLEncoding.EncodeToString(guidBuff))
}

func BuildDockerImageStartByHTTPReq(worker string, job *Job) {

	dockerClient, _ := NewDockerClient(worker, nil)

	reader, err := dockerClient.BuildImage(&(job.ImageConfig))
	if err != nil {
		log.Println("[ErrorInfo]", err.Error())

	} else if reader != nil {
		buf := make([]byte, 4096)

		for {
			n, err := reader.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}

			dockerClient.PushImage(&(job.ImageConfig))

			if 0 == n {
				err = models.PushMsgToList("buildLog:"+job.Tag, "bye")
				if err != nil {
					log.Println("[ErrorInfo]", err.Error())
				}

				err = models.PublishMsg("buildLog:"+job.Tag, "bye")
				if err != nil {
					log.Println("[ErrorInfo]", err.Error())
				}
				finishJob(worker)
				break
			}

			err = models.PushMsgToList("buildLog:"+job.Tag, string(buf[:n]))
			if err != nil {
				log.Println("[ErrorInfo]", err.Error())
			}

			err = models.PublishMsg("buildLog:"+job.Tag, string(buf[:n]))
			if err != nil {
				log.Println("[ErrorInfo]", err.Error())
			}
		}
	} else {

		log.Println("[ErrorInfo]", err.Error())
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
		var in io.Reader
		if job.Mode == "dockerfile" {
			in = tarDockerFile(job.Context)
		} else if job.Mode == "archive" {
			in = readArchive(job.Context)
		} else {
			log.Printf("[ErrorInfo] : %v\n", errors.New("Wrong mode, required mode exactly."))
		}

		job.ImageConfig.Context = in

		go BuildDockerImageStartByHTTPReq(worker, job)
	}
}

func readArchive(archivePath string) io.Reader {
	f, err := os.Open(archivePath)
	if err != nil {
		log.Println("[ErrorInfo]", err)
	}
	defer f.Close()
	buf, _ := ioutil.ReadAll(f)

	return bytes.NewReader(buf)
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

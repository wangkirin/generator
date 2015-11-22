package models

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/containerops/wrench/db"
	"github.com/satori/go.uuid"

	. "github.com/containerops/generator/modules/build"
)

const (
	GEN_JOB = "GEN_JOB"

	UNSTART    = "UNSTART"
	BUILDING   = "BUILDING"
	BUILDERROR = "BUILD_ERROR"
	BUILDDONE  = "BUILD_DONE"
	PUSHING    = "PUSHING"
	PUSHERROR  = "PUSH_ERROR"
	PUSHED     = "PUSHED"
	DEL        = "DEL"

	SHOULD_REBUILD_STATUS = UNSTART + "," + BUILDING + "," + BUILDERROR
	SHOULD_REPUSH_STATUS  = BUILDDONE + "," + PUSHING + "," + PUSHERROR
	SHOULD_DEL_STATUS     = PUSHED + "," + DEL

	BUILDLOGSTR = "BuildLog_%s"
	PUSHLOGSTR  = "PushLog_%s"
)

type Job struct {
	UUID           string `json:"uuid"`
	Repo           string `json:"repo"`
	Tag            string `json:"tag"`
	Host           string `json:"host"`
	Port           string `json:"port"`
	Dockerfile     string `json:"dockerfile"`
	Status         string `json:"status"`
	BuildLog       string `json:"build_log"`
	PushLog        string `json:"push_log"`
	BuildImgConfig string `json:"build_img_config"`
}

var (
	JobsChans = make(chan *Job, 65535)
)

func InitJob() {
	keys, err := db.Client.HKeys(GEN_JOB).Result()
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		index := strings.LastIndex(key, ":")
		if index > 0 {
			repo := key[0:index]
			tag := key[index+1 : len(key)]

			jobStrs, err := db.Client.LRange(GEN_JOB+"_"+repo+"_"+tag, int64(0), int64(-1)).Result()
			if err != nil {
				log.Println("[errorInfo]: init job error when get info from db:", err.Error())
			}
			for _, jobStr := range jobStrs {
				job := new(Job)
				err := json.Unmarshal([]byte(jobStr), job)
				if err != nil {
					log.Println("[errorInfo]:init job error when get info from db:", err.Error())
					continue
				}
				JobsChans <- job
			}
		} else {
			panic(errors.New("init job by db error!"))
		}
	}
	go startJob()
}

func (j *Job) Add(repo, tag, dockerfile, buildImgConfig string) (string, error) {
	uid := uuid.NewV1().String()
	job := &Job{
		UUID:           uid,
		Repo:           repo,
		Tag:            tag,
		Dockerfile:     dockerfile,
		Status:         UNSTART,
		BuildLog:       fmt.Sprintf(BUILDLOGSTR, uid),
		PushLog:        fmt.Sprintf(PUSHLOGSTR, uid),
		BuildImgConfig: buildImgConfig,
	}

	jobByte, err := json.Marshal(job)
	if err != nil {
		log.Println("[errorInfo]:Add job fail:", err.Error())
		return "", err
	}

	_, err = db.Client.HSet(GEN_JOB, repo+":"+tag, repo+":"+tag).Result()
	if err != nil {
		return "", err
	}

	_, err = db.Client.LPush(GEN_JOB+"_"+repo+"_"+tag, string(jobByte)).Result()
	if err != nil {
		return "", err
	}

	JobsChans <- job
	return job.UUID, nil
}

func startJob() {
	for {
		job := <-JobsChans
		if strings.Contains(SHOULD_DEL_STATUS, job.Status) {
			go doDel(job)
		} else {
			deamon := <-DeamonChans
			if strings.Contains(SHOULD_REBUILD_STATUS, job.Status) {
				go doBuild(job, deamon)
			} else if strings.Contains(SHOULD_REPUSH_STATUS, job.Status) {
				go doPush(job, deamon)
			} else {
				log.Println("start an unrecognized job,job id is", job.UUID)
			}
		}
	}
}

func doBuild(job *Job, deamon *Deamon) {

	doDel(job)
	job.Host = deamon.Host
	job.Port = deamon.Port
	job.Status = BUILDING
	doAdd(job)

	dockerClient, _ := NewDockerClient(deamon.Host+":"+deamon.Port, nil)

	reader, err := tarDockerFile(job.Dockerfile)

	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	buildImageConfig := new(BuildImage)
	err = json.Unmarshal([]byte(job.BuildImgConfig), buildImageConfig)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	buildImageConfig.Context = reader
	buildLogReader, err := dockerClient.BuildImage(buildImageConfig)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	buf := make([]byte, 4096)

	for {
		n, err := buildLogReader.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalln(err)
			panic(err)
		}

		if n == 0 {
			db.Client.RPush(fmt.Sprintf(BUILDLOGSTR, job.UUID), "build end,please read push log")
			db.Client.RPush(fmt.Sprintf(BUILDLOGSTR, job.UUID), "bye")
			doDel(job)
			job.Status = BUILDDONE
			doAdd(job)
			JobsChans <- job
			break
		}

		_, err = db.Client.RPush(job.BuildLog, string(buf[:n])).Result()
		if err != nil {
			log.Fatalln(err)
			panic(err)
		}

	}
}

func doPush(job *Job, deamon *Deamon) {
	defer postEnd(job, deamon)

	doDel(job)
	job.Status = PUSHING
	doAdd(job)

	dockerClient, _ := NewDockerClient(deamon.Host+":"+deamon.Port, nil)

	reader, err := tarDockerFile(job.Dockerfile)

	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	buildImageConfig := new(BuildImage)
	buildImageConfig.Context = reader

	//start push build image
	pushLogReader, err := dockerClient.PushImage(job.Repo, job.Tag, buildImageConfig)
	pBuf := make([]byte, 4096)

	for {
		m, err := pushLogReader.Read(pBuf)
		if err != nil && err != io.EOF {
			log.Fatalln(err)
			panic(err)
		}
		if 0 == m {
			db.Client.RPush(job.PushLog, "pushed")
			db.Client.RPush(job.PushLog, "bye")
			break
		}

		_, err = db.Client.RPush(job.PushLog, string(pBuf[:m])).Result()
		if err != nil {
			log.Fatalln(err)
			panic(err)
		}
	}
}

func postEnd(job *Job, deamon *Deamon) {
	deamon.restore(deamon.Host, deamon.Port)
	doDel(job)
	job.Status = PUSHED
	doAdd(job)
	JobsChans <- job
}

func doDel(job *Job) {
	jobByte, err := json.Marshal(job)
	_, err = db.Client.LRem(GEN_JOB+"_"+job.Repo+"_"+job.Tag, int64(0), string(jobByte)).Result()
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	count, err := db.Client.LLen(GEN_JOB + "_" + job.Repo + "_" + job.Tag).Result()
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	if count == int64(0) {
		db.Client.HDel(GEN_JOB, job.Repo+":"+job.Tag)
	}
}

func doAdd(job *Job) {
	jobByte, err := json.Marshal(job)
	_, err = db.Client.HSet(GEN_JOB, job.Repo+":"+job.Tag, job.Repo+":"+job.Tag).Result()
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	_, err = db.Client.LPush(GEN_JOB+"_"+job.Repo+"_"+job.Tag, string(jobByte)).Result()
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
}

func GetLogInfo(key string, start, stop int64) (logs []string, err error) {
	logs, err = db.Client.LRange(key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetAllJobs() ([]string, error) {
	repos, err := db.Client.HGetAll(GEN_JOB).Result()
	result := make([]string, 0)
	if err != nil {
		return result, err
	}
	for _, v := range repos {
		index := strings.LastIndex(v, ":")
		if index > 0 {
			repo := v[0:index]
			tag := v[index+1 : len(v)]

			jobs, err := db.Client.LRange(GEN_JOB+"_"+repo+"_"+tag, int64(0), int64(-1)).Result()
			if err != nil {
				return nil, err
			}
			result = append(result, jobs...)
		}
	}
	return result, nil
}

func tarDockerFile(dockerfile string) (reader io.Reader, err error) {

	dockerfileBytes, err := base64.StdEncoding.DecodeString(dockerfile)

	if err != nil {
		log.Fatalln(err)
		return nil, err
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
			return nil, err
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
			return nil, err
		}
	}

	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalln(err)
		return nil, err
	}

	tarReader := bytes.NewReader(buf.Bytes())

	return tarReader, nil
}

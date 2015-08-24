package modules

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_DockerBuild(t *testing.T) {
	Convey("Test Docker Build", t, func() {

		testDockerUrl := "tcp://192.168.19.53:9999"
		testTarPath := "test.tar"

		// Init the client
		docker, _ := NewDockerClient(testDockerUrl, nil)

		// Build a docker image
		// some.tar contains the build context (Dockerfile any any files it needs to add/copy)
		dockerBuildContext, err := os.Open(testTarPath)
		defer dockerBuildContext.Close()

		buildImageConfig := &BuildImage{
			Context:        dockerBuildContext,
			RepoName:       "your_image_name",
			SuppressOutput: false,
		}

		reader, err := docker.BuildImage(buildImageConfig)
		if err != nil {
			log.Fatal(err)
		}

		buf := make([]byte, 4096)
		for {

			n, err := reader.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if 0 == n {
				break
			}

			// find \u001b[91m
			reg, _ := regexp.Compile(`\\u001b\[91m`)
			jsonText := string(buf[:n])
			if reg.MatchString(jsonText) {
				reg, err := regexp.Compile(`[\d]+\%`)

				if err == nil {
					percentText := reg.FindString(jsonText)
					if len(percentText) > 0 {
						fmt.Print(percentText, "|")
					}
				}
			} else {
				fmt.Println(jsonText)
				// chunks=append(chunks,buf[:n]...)

			}
		}
	})
}

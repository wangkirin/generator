package modules

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	APIVersion = "v1.21"
)

var (
	ErrNotFound = errors.New("Not found")

	defaultTimeout = 30 * time.Second
)

type DockerClient struct {
	URL        *url.URL
	HTTPClient *http.Client
	TLSConfig  *tls.Config
}

type Error struct {
	StatusCode int
	Status     string
	msg        string
}

type Client interface {
	BuildImage(image *BuildImage) (io.ReadCloser, error)
}

// BuildImage : config for buildImage()
// Note:because it is recepted as params from restful call and to be setted to url.values{},
// The type of them are all string
type BuildImage struct {
	Dockerfile     string
	Context        io.Reader
	RemoteURL      string
	RepoName       string
	SuppressOutput string
	NoCache        string
	Remove         string
	ForceRemove    string
	Pull           string
	Memory         string
	MemorySwap     string
	CpuShares      string
	CpuPeriod      string
	CpuQuota       string
	CpuSetCpus     string
	CpuSetMems     string
	CgroupParent   string
}

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.msg)
}

func NewDockerClient(daemonUrl string, tlsConfig *tls.Config) (*DockerClient, error) {
	return NewDockerClientTimeout(daemonUrl, tlsConfig, time.Duration(defaultTimeout))
}

func NewDockerClientTimeout(daemonUrl string, tlsConfig *tls.Config, timeout time.Duration) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Scheme == "tcp" {
		if tlsConfig == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	httpClient := newHTTPClient(u, tlsConfig, timeout)
	return &DockerClient{u, httpClient, tlsConfig}, nil
}

func (client *DockerClient) doStreamRequest(method string, path string, in io.Reader, headers map[string]string) (io.ReadCloser, error) {
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, client.URL.String()+path, in)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
	}

	resp, err := client.HTTPClient.Do(req)
	if resp.StatusCode == 200 {
		return resp.Body, nil
	} else if resp.StatusCode == 400 || resp.StatusCode == 404 || resp.StatusCode == 409 || resp.StatusCode == 500 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, msg: string(data)}
	}
	return resp.Body, errors.New("Unknown err from server")
}

// Todo : need delete build temp containers
func (client *DockerClient) BuildImage(image *BuildImage) (io.ReadCloser, error) {

	v := url.Values{}

	if image.Dockerfile != "" {
		v.Set("dockerfile", image.Dockerfile)
	}

	if image.RepoName != "" {
		v.Set("t", image.RepoName)
	} else {
		return nil, errors.New("Empty reponame, required params")
	}

	if image.RemoteURL != "" {
		v.Set("remote", image.RemoteURL)
	}

	if image.NoCache != "" {
		v.Set("nocache", image.NoCache)
	}

	if image.Pull != "" {
		v.Set("pull", image.Pull)
	}

	if image.Remove != "" {
		v.Set("rm", image.Remove)
	}

	if image.ForceRemove != "" {
		v.Set("forcerm", image.ForceRemove)
	}

	if image.SuppressOutput != "" {
		v.Set("q", image.SuppressOutput)
	}

	if image.Memory != "" {
		v.Set("memory", image.Memory)
	}
	if image.MemorySwap != "" {
		v.Set("memoryswap", image.MemorySwap)
	}
	if image.CpuShares != "" {
		v.Set("cpushares", image.CpuShares)
	}

	if image.CpuPeriod != "" {
		v.Set("cpuperiod", image.CpuPeriod)
	}

	if image.CpuQuota != "" {
		v.Set("cpuquota", image.CpuQuota)
	}

	if image.CpuSetCpus != "" {
		v.Set("cpusetcpus", image.CpuSetCpus)
	}

	if image.CpuSetMems != "" {
		v.Set("cpusetmems", image.CpuSetMems)
	}

	if image.CgroupParent != "" {
		v.Set("cgroupparent", image.CgroupParent)
	}

	headers := make(map[string]string)
	if image.Context != nil {
		headers["Content-Type"] = "application/tar"
	}

	uri := fmt.Sprintf("/%s/build?%s", APIVersion, v.Encode())
	return client.doStreamRequest("POST", uri, image.Context, headers)
}

func (client *DockerClient) PushImage(image *BuildImage) (io.ReadCloser, error) {

	headers := make(map[string]string)
	headers["X-Registry-Auth"] = "Og=="
	uri := fmt.Sprintf("/images/%s/push", image.RepoName)
	return client.doStreamRequest("POST", uri, image.Context, headers)
}

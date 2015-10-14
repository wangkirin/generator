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
	"strconv"
	"strings"
	"time"
)

const (
	APIVersion = "v1.20"
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

type BuildImage struct {
	DockerfileName string
	Context        io.Reader
	RemoteURL      string
	RepoName       string
	SuppressOutput bool
	NoCache        bool
	Remove         bool
	ForceRemove    bool
	Pull           bool
	Memory         int64
	MemorySwap     int64
	CpuShares      int64
	CpuPeriod      int64
	CpuQuota       int64
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
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") && client.TLSConfig == nil {
			return nil, fmt.Errorf("%v. Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, msg: string(data)}
	}

	return resp.Body, nil
}

// Todo : need delete build temp containers
func (client *DockerClient) BuildImage(image *BuildImage) (io.ReadCloser, error) {
	v := url.Values{}

	if image.DockerfileName != "" {
		v.Set("dockerfile", image.DockerfileName)
	}
	if image.RepoName != "" {
		v.Set("t", image.RepoName)
	}
	if image.RemoteURL != "" {
		v.Set("remote", image.RemoteURL)
	}
	if image.NoCache {
		v.Set("nocache", "1")
	}
	if image.Pull {
		v.Set("pull", "1")
	}
	if image.Remove {
		v.Set("rm", "1")
	} else {
		v.Set("rm", "0")
	}
	if image.ForceRemove {
		v.Set("forcerm", "1")
	}
	if image.SuppressOutput {
		v.Set("q", "1")
	}

	v.Set("memory", strconv.FormatInt(image.Memory, 10))
	v.Set("memswap", strconv.FormatInt(image.MemorySwap, 10))
	v.Set("cpushares", strconv.FormatInt(image.CpuShares, 10))
	v.Set("cpuperiod", strconv.FormatInt(image.CpuPeriod, 10))
	v.Set("cpuquota", strconv.FormatInt(image.CpuQuota, 10))
	v.Set("cpusetcpus", image.CpuSetCpus)
	v.Set("cpusetmems", image.CpuSetMems)
	v.Set("cgroupparent", image.CgroupParent)

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

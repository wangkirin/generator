# Generator - A docker image builder

Generator is a conventional Docker image builder that simply accepts webhooks from any Github repository, builds an image for that repository, and pushes it to the supplied registry.

## How it works

* Generator receives a build request, for example, via a GitHub commit webhook.
* Generator builds and tags the resulting image.
* Generator then pushes the image to the supplied Docker registry such as dockyard.

## Usage

### Compile

Compile is as simple as:

```bash
# create a 'github.com/containerops' directory in your GOPATH/src
cd github.com/containerops
git clone https://github.com/containerops/generator
cd generator
go build
```

### Configuration

Before using `generator` service, some prerequisites should be done first.

#### Deploy dockyard

As the default registry for generator, dockyard should be deployed starting `generator` service. You can follow the [instruction](https://github.com/containerops/dockyard#try-it-out) to complete the operation.

#### Deploy docker build resource pool

Add a config file named `pool.json` at `generator/conf` directory before starting `generator` service. Below is a `pool.json` example:

```
{
   "docker":[
      {
         "url":"127.0.0.1",
         "port":"19000"
      },
      {
         "url":"127.0.0.1",
         "port":"19100"
      }
   ]
}
```

Run docker daemon corresponding to the pool.json

```bash
$ docker daemon --insecure-registry=containerops.me -iptables=false -H :19000 -g /opt/docker-data/19000 &
$ docker daemon --insecure-registry=containerops.me -iptables=false -H :19100 -g /opt/docker-data/19100 &
```

#### Config runtime parameters
Add a config file named `runtime.conf` at `generator/conf` directory before starting `generator` service. Below is a `runtime.conf` example:

```
runmode = dev
listenmode = http
httpscertfile = cert/containerops/containerops.crt
httpskeyfile = cert/containerops/containerops.key

[log]
filepath = log/containerops-log

[db]
uri = localhost:6379
passwd = containerops
db = 8

[generator]
genurl = containerops.me
dockerfilepath  = /tmp
```

### Run

Start `generator` service:

```bash
$ ./generator web &
```

### Examples

#### Triggering a build by dockerfile

* Preparing Dockerfile
```
FROM       ubuntu

MAINTAINER Cheng Tiesheng <chengtiesheng@huawei.com>

ENV TERM xterm

EXPOSE 80
```

* Encoding Dockerfile by Base64
```
RlJPTSAgICAgICB1YnVudHUNCg0KTUFJTlRBSU5FUiBDaGVuZyBUaWVzaGVuZyA8Y2hlbmd0aWVzaGVuZ0BodWF3ZWkuY29tPg0KDQpFTlYgVEVSTSB4dGVybQ0KDQpFWFBPU0UgODA=
```

* Use Web broswer to trigger REST API call

```
http://localhost:8080/b1/build?mode=dockerfile&imagename=containerops.me/cts/example&context=RlJPTSAgICAgICB1YnVudHUNCg0KTUFJTlRBSU5FUiBDaGVuZyBUaWVzaGVuZyA8Y2hlbmd0aWVzaGVuZ0BodWF3ZWkuY29tPg0KDQpFTlYgVEVSTSB4dGVybQ0KDQpFWFBPU0UgODA=
```

Then you will ge a job id such as:
```
f96ab2530d17c08716d2850d17c08729
```

* Following the progress of a build

You can get the build log by using web socket:
```
ws://localhost:8080/b1/build/log/ws/f96ab2530d17c08716d2850d17c08729
```

* Get docker image from dockyard

Now, the docker image containerops.me/cts/example has been pushed to dockyard, you can get it by using docker pull:
```
docker pull containerops.me/cts/example
```
     
#### trigger a build by a archive     
    
* Prepare archive including dockerfile
     
Using [hello-world](https://github.com/docker-library/hello-world/tree/b7a78b7ccca62cc478919b101f3ab1334899df2b) as a example    
Download it from github  
```          
cd hello-world     
tar cvf hello.tar * 
mkdir /tmp
cp hello.tar /tmp/
  
```    
* Use web browser to trigger REST API CALL     
```    
http://localhost:8080/b1/build?mode=archive&imagename=containerops.me/cts/example&context=/tmp/hello.tar   
```     
* Debug and get logs   
      
Debug and get logs as the way in topic 'trigger a build by dockerfile'   



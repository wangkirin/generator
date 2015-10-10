# Generator
===========

##send Dockerfile

```
FROM       ubuntu
MAINTAINER MengFanliang <mengfanliang@huawei.com>

ENV TZ "Asia/Shanghai"
ENV TERM xterm

RUN apt-get install wget -y

RUN wget https://get.docker.com/builds/Darwin/x86_64/docker-latest


EXPOSE 22
```

## conf/pool.json

```
{
   "docker":[
      {
         "url":"192.168.199.102",
         "port":"9999"
      },
      {
         "url":"192.168.199.103",
         "port":"9999"
      }
   ]
}
```

## WebSocket Get Build log URL 

```
	Method : GET
	URL : ws://192.168.199.10:8080/wsbuildlog
	
	send json info
	{"id":"a95d9886304920ad3437aeb2c7cea2a3"}

```
## HTTP Get Build log URL 
```
	Method : POST

	URL : http://192.168.199.10:8080/httpbuildlog
	
	Param : logid=cd48ff1786c2dd8d86172662b07a8103&count=1

	Return : {"status":"Downloading",.......}
```

## Send Build Dockerfile

```
	Method : POST
	
	URL : http://192.168.199.10:8080/httpbuild
	
	Param : imagename=containerops.me:5000/fsk/hw2ubuntu:15.04&dockerfile=(BASE64DockerFile)


	Return{LogID} : 9a1df7f8833cbb706f45a00882e200f7
```

## runtime.conf

```
runmode = dev
listenmode = https
httpscertfile = cert/containerops/containerops.crt
httpskeyfile = cert/containerops/containerops.key

[log]
filepath = log/containerops-log

[db]
uri = localhost:6379
passwd = containerops
db = 8

[generator]
genurl = 192.168.19.112:9999
dockerfilepath  = /tmp
```

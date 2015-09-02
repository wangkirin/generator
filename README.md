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

## websocket send build dockerfile json

```
{
  "name":"build-test",
  "dockerfile":"IyBET0NLRVItVkVSU0lPTiAgICAxLjguMQoKRlJPTSAgICAgICB1YnVudHUKTUFJTlRBSU5FUiBNZW5nRmF"
}
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

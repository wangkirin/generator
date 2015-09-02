# Generator
===========

##send Dockerfile

FROM       ubuntu
MAINTAINER MengFanliang <mengfanliang@huawei.com>

ENV TZ "Asia/Shanghai"
ENV TERM xterm

RUN apt-get install wget -y

RUN wget https://get.docker.com/builds/Darwin/x86_64/docker-latest


EXPOSE 22

## websocket send build dockerfile json

{"name":"build-test","dockerfile":"IyBET0NLRVItVkVSU0lPTiAgICAxLjguMQoKRlJPTSAgICAgICB1YnVudHUKTUFJTlRBSU5FUiBNZW5nRmFubGlhbmcgPG1lbmdmYW5saWFuZ0BodWF3ZWkuY29tPgoKRU5WIFRaICJBc2lhL1NoYW5naGFpIgpFTlYgVEVSTSB4dGVybQoKUlVOIGFwdC1nZXQgaW5zdGFsbCB3Z2V0IC15CgpSVU4gd2dldCBodHRwczovL2dldC5kb2NrZXIuY29tL2J1aWxkcy9EYXJ3aW4veDg2XzY0L2RvY2tlci1sYXRlc3QKCgpFWFBPU0UgMjIK"}

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

#### Deploy redis service

* Download and install redis package:

```bash
$ curl -sL http://download.redis.io/releases/redis-stable.tar.gz | tar xzf - -C /tmp --strip 1
$ make -C /tmp
$ make -C /tmp install
$ mkdir /var/lib/redis
```

* Add redis.conf at `etc/redis` directory:

```
daemonize yes
pidfile /var/run/redis.pid
port 6379
tcp-backlog 511
bind 127.0.0.1
timeout 0
tcp-keepalive 0
loglevel notice
logfile ""
databases 16
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /var/lib/redis
slave-serve-stale-data yes
slave-read-only yes
repl-disable-tcp-nodelay no
slave-priority 100
requirepass containerops
appendonly no
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-entries 512
list-max-ziplist-value 64
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes
```

* Start redis service:

```bash
$ redis-server /etc/redis/redis.conf
```

* Test redis service:

```bash
$ redis-cli
127.0.0.1:6379> auth containerops
OK
127.0.0.1:6379> ping
PONG
127.0.0.1:6379> 
```

#### Deploy docker build resource pool

* Add a config file named `pool.json` at `generator/conf` directory before starting `generator` service. Below is a `pool.json` example:

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

* Run docker daemon corresponding to the pool.json

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

#### Triggering a build

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
http://localhost:8080/b1/build?imagename=containerops.me/cts/example&dockerfile=RlJPTSAgICAgICB1YnVudHUNCg0KTUFJTlRBSU5FUiBDaGVuZyBUaWVzaGVuZyA8Y2hlbmd0aWVzaGVuZ0BodWF3ZWkuY29tPg0KDQpFTlYgVEVSTSB4dGVybQ0KDQpFWFBPU0UgODA=
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

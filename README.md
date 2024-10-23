# How to run
Requirements:
- Linux host
- Docker
- [gVisor](https://gvisor.dev/docs/user_guide/install/)

```shell
$ git pull https://github.com/iklabib/Kerat.git
# sorry, no public image yet, so you have to build it yourself
$ docker build . -t kerat:collections -f containerfiles/collections.Dockerfile
$ docker run --runtime=runsc -p 3145:3145 -v /var/run/docker.sock:/var/run/docker.sock -v /app/config.yaml:./config.yaml -it kerat:collections
```
If you don't have Linux machine and just want to test things out, remove `runtime` from `config.yaml` to fallback to Docker's default runtime.

# How it works?
We have a compiler container that receive source code, compile them to executable binary, and run said executable in another container. No brainer.

# Motivation
I slapped gvisor to container and call it a sandbox. In nutshell, just like Go Playground does. 

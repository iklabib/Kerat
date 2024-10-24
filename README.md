# Kerat
I slapped gvisor to container and call it a sandbox. In nutshell, just like Go Playground does. 

## How it works?
We have a compiler container that receive source code, compile them to executable binary, and run said executable in another container. No brainer.

## How to run
Requirements:
- Linux host
- Docker
- [gVisor](https://gvisor.dev/docs/user_guide/install/)

We want to able to use gvisor but give it a bit of slack so it able to use docker socket. Skip this step if you want to use Docker's default runtime.

Copy `runsc-uds` to your`/etc/docker/daemon.json` and reload docker `sudo systemctl reload docker.service`.
```json
{
    "runtimes": {
        "runsc": {
            "path": "/usr/local/bin/runsc"
       },
        "runsc-uds": {
          "path": "/usr/local/bin/runsc",
          "runtimeArgs": [
            "--host-uds=open"
        ]
       }
    }
}
```

```shell
$ git pull https://github.com/iklabib/Kerat.git
# sorry, no public image yet, so you have to build it yourself
$ docker build . -t kerat:engine
$ docker run --runtime=runsc-uds -p 3145:3145 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --mount type=bind,source="$(pwd)"/config.yaml,target=/app/config.yaml \
    -it kerat:engine
```
If you don't have Linux machine and just want to test things out, remove `runtime` from `config.yaml` to fallback to Docker's default runtime.
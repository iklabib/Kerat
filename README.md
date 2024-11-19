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
$ curl -sSL https://raw.githubusercontent.com/iklabib/Kerat/refs/heads/main/build.sh -o build.sh
$ chmod +x build.sh
$ ./build.sh pull
$ docker run --runtime=runsc-uds -p 31415:31415 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -it kerat:engine
```
or you want to build them yourself. It has amd64 and arm64 support.

```shell
$ git pull https://github.com/iklabib/Kerat.git
$ cd Kerat
$ ./build.sh all amd64 # or arm64
```
If you don't have Linux machine and just want to test things out, remove `runtime` from `config.yaml` to fallback to Docker's default runtime.

Request sample (not exactly user friendly, there is a reason for that).
```shell
$ curl --request POST \
  --url http://127.0.0.1:31415/submit \
  --header 'content-type: application/json' \
  --data '{
  	"exercise_id": "dummy",
  	"type": "python",
  	"source": {
      "src_test": [
          {
              "filename": "test_example.py",
              "src": "import unittest\n\nclass TestExample(unittest.TestCase):\n    def test_addition(self):\n        self.assertEqual(1+1, 2)"
          }
      ],
      "src": [
          {
              "filename": "example.py",
              "src": "def add(a, b):\n    return a + b"
          }
      ]
    }
}'

# output sample
# {
#   "message": "",
#   "success": true,
#   "output": [
#     {
#       "passed": true,
#       "name": "test_addition",
#       "stack_trace": ""
#     }
#   ]
# }
```
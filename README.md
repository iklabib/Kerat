# Kerat
I slapped gvisor to container and call it a sandbox. In nutshell, just like Go Playground does. 

## How it works?
We have a compiler container that receive source code, compile them to executable binary, and run said executable in another container. No brainer.

## How to run
Requirements:
- Linux host
- Docker
- [gVisor](https://gvisor.dev/docs/user_guide/install/)

Kerat spawn another container, so it need host's docker socket and pulling runtime images.

```shell
$ curl -sSL https://raw.githubusercontent.com/iklabib/Kerat/refs/heads/main/build.sh -o build.sh
$ chmod +x build.sh
$ ./build.sh pull
$ docker run -p 31415:31415 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -it iklabib/kerat:engine
```
You can build the images yourself if you want to, with amd64 and arm64 support.

```bash
$ git pull https://github.com/iklabib/Kerat.git
$ cd Kerat
$ ./build.sh all amd64 # or arm64
```

Request sample (not exactly user friendly, there is a reason for that).
```bash
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
#  "success": true,
#  "message": "",
#  "output": [
#    {
#      "passed": false,
#      "name": "test_addition",
#      "message": "AssertionError: 4 != 2\n",
#      "stack_trace": ""
#    }
#  ],
#  "metrics": {
#    "exit_code": 0,
#    "wall_time": 0.63,
#    "cpu_time": 35539000,
#    "memory": 2375680
#  }
# }
```

## Running Kerat:engine with gVisor
Kerat:engine is the container that compiles source codes and spawn container to run them. It need access to host's docker socket, this is blocked by default by gVisor. Here is how to get around the issue.

Add `runsc-uds` lines to your `/etc/docker/daemon.json` and reload docker `sudo systemctl reload docker.service`.
```bash
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

Now you can use `runsc-uds` as docker runtime.

```bash
$ docker run --runtime=runsc-uds -p 31415:31415 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -it kerat:engine
```
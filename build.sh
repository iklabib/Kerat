#!/bin/bash

ENGINE=podman
case "$1" in
alpine)
  $ENGINE build . -t kerat:alpine -f containerfiles/alpine.Dockerfile
  ;;

csharp)
  $ENGINE build . -t kerat:csharp -f containerfiles/csharp.Dockerfile
  ;;
esac

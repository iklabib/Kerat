#!/bin/bash

ENGINE=docker
case "$1" in
alpine)
  $ENGINE build . -t kerat:alpine -f containerfiles/alpine.Dockerfile
  ;;

csharp)
  $ENGINE build . -t kerat:csharp -f containerfiles/csharp.Dockerfile
  ;;

csharp)
  $ENGINE build . -t kerat:python -f containerfiles/python.Dockerfile
  ;;
esac

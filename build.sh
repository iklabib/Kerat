#!/bin/bash

ENGINE=docker
case "$1" in
alpine)
  $ENGINE build . -t kerat:alpine -f containerfiles/alpine.Dockerfile
  ;;

csharp)
  $ENGINE build . -t kerat:csharp -f containerfiles/csharp.Dockerfile
  ;;

python)
  $ENGINE build . -t kerat:python -f containerfiles/python.Dockerfile
  ;;

java)
  $ENGINE build . -t kerat:java -f containerfiles/java.Dockerfile
  ;;

kotlin)
  $ENGINE build . -t kerat:kotlin -f containerfiles/kotlin.Dockerfile
  ;;

base)
  $ENGINE build . -t kerat:base -f containerfiles/base.Dockerfile
  ;;

kerat)
  $ENGINE build . -t kerat:engine -f Dockerfile
  ;;
esac

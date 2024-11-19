#!/bin/bash

ENGINE=docker
ARCH=${2:-"amd64"} 

if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" ]]; then
  echo "Error: Supported architectures are 'amd64' or 'arm64'"
  exit 1
fi

build() {
  $ENGINE buildx build . -t iklabib/kerat:$1 -f ${2:-containerfiles/$1.Dockerfile} --build-arg ARCH=$ARCH --platform linux/$ARCH 
}

build_all() {
  declare -A targets=(
    ["box"]="containerfiles/box.Dockerfile"
    ["box-alpine"]="containerfiles/box-alpine.Dockerfile"
    ["python"]="containerfiles/python.Dockerfile"
    ["kerat"]="Dockerfile"
  )

  for target in "${!targets[@]}"; do
    echo "Building $target..."
    build "$target" "${targets[$target]}"
  done
}

pull_all() {
  # List of all targets to pull
  images=(
    "iklabib/kerat:box"
    "iklabib/kerat:box-alpine"
    "iklabib/kerat:python"
    "iklabib/kerat:engine"
  )

  for image in "${images[@]}"; do
    echo "Pulling $image..."
    $ENGINE pull "$image"
  done
}

case "$1" in
  box) build box containerfiles/box.Dockerfile ;;
  box-alpine) build box-alpine containerfiles/box-alpine.Dockerfile ;;
  python) build python containerfiles/python.Dockerfile ;;
  kerat) build kerat Dockerfile ;;
  all) build_all ;;
  pull) pull_all ;;
  *) echo "Usage: $0 {box|box-alpine|python|kerat|all} [ARCH]"
     echo "ARCH must be 'amd64' or 'arm64'." ;;
esac

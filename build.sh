#!/bin/bash

ENGINE=docker
ARCH=${2:-"amd64"}

if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" && "$ARCH" != "all" ]]; then
  echo "Error: Supported architectures are 'amd64', 'arm64' or 'all'"
  exit 1
fi

build() {
  if [[ "$ARCH" == "all" ]]; then
    PLATFORMS="linux/amd64,linux/arm64"
  else
    PLATFORMS="linux/$ARCH"
  fi

  $ENGINE buildx build . -t iklabib/kerat:$1 -f ${2:-containerfiles/$1.Dockerfile} --platform $PLATFORMS --push
}

build_all() {
  declare -A targets=(
    ["box"]="containerfiles/box.Dockerfile"
    ["box-alpine"]="containerfiles/box-alpine.Dockerfile"
    ["python"]="containerfiles/python.Dockerfile"
    ["engine"]="Dockerfile"
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
  engine) build engine Dockerfile ;;
  all) build_all ;;
  pull) pull_all ;;
  *) echo "Usage: $0 {box|box-alpine|python|engine|all} [ARCH]"
     echo "ARCH must be 'amd64', 'arm64', or 'all' for multi-arch builds." ;;
esac

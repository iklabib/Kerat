#!/bin/bash

ENGINE="docker"
ARCH="${2:-amd64}"

if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" && "$ARCH" != "all" ]]; then
    echo "Error: Supported architectures are 'amd64', 'arm64', or 'all'"
    exit 1
fi

build() {
    local target="$1"
    local dockerfile="$2"
    local push="$3"

    if [[ "$ARCH" == "all" ]]; then
        platforms="linux/amd64,linux/arm64"
    else
        platforms="linux/$ARCH"
    fi

    dockerfile="${dockerfile:-containerfiles/$target.Dockerfile}"
    command="$ENGINE buildx build . -t iklabib/kerat:$target -f $dockerfile --platform $platforms"
    if $push; then
        command="$command --push"
    else
        command="$command --load"
    fi

    eval $command
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
        build "$target" "${targets[$target]}" "$push"
    done
}

pull_all() {
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

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 {box|box-alpine|python|engine|all} [ARCH]"
    echo "ARCH must be 'amd64', 'arm64', or 'all' for multi-arch builds."
    exit 1
fi

push=false
if [[ $# -gt 3 && "$3" == "--push" ]]; then
    push=true
fi

case "$1" in
    "box")
        build "box" "containerfiles/box.Dockerfile" "$push"
        ;;
    "box-alpine")
        build "box-alpine" "containerfiles/box-alpine.Dockerfile" "$push"
        ;;
    "python")
        build "python" "containerfiles/python.Dockerfile" "$push"
        ;;
    "engine")
        build "engine" "Dockerfile" "$push"
        ;;
    "all")
        build_all
        ;;
    "pull")
        pull_all
        ;;
    *)
        echo "Usage: $0 {box|box-alpine|python|engine|all}"
        echo "ARCH must be 'amd64', 'arm64', or 'all' for multi-arch builds."
        exit 1
        ;;
esac

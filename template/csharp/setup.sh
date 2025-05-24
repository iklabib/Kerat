#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

DOTNET_VERSION=${1:-"8.0.404"}
ARCH=${2:-"amd64"}  # Default to "amd64" if not provided

if [ "$ARCH" = "amd64" ]; then
    ARCH="x64"
fi

DOTNET_DIR="$HOME/dotnet"
mkdir -p "$DOTNET_DIR"

# Download and extract .NET SDK
TARBALL="dotnet-sdk-$DOTNET_VERSION-linux-$ARCH.tar.gz"
aria2c -x 16 -s 16 "https://dotnetcli.azureedge.net/dotnet/Sdk/$DOTNET_VERSION/$TARBALL"
tar -xf "$TARBALL" -C "$DOTNET_DIR"
rm "$TARBALL"

# Update PATH for the current script
export PATH="$DOTNET_DIR:$PATH"

# Restore and publish the .NET project
dotnet restore "$SCRIPT_DIR/box.csproj"
dotnet publish -o "$SCRIPT_DIR/output" "$SCRIPT_DIR/box.csproj"

rm -rf $SCRIPT_DIR/output
rm $SCRIPT_DIR/setup.sh

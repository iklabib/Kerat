#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DOTNET_VERSION=$1
mkdir ~/dotnet
aria2c -x 16 -s 16 "https://dotnetcli.azureedge.net/dotnet/Sdk/8.0.404/dotnet-sdk-$DOTNET_VERSION-linux-x64.tar.gz"
tar -xf "dotnet-sdk-$DOTNET_VERSION-linux-x64.tar.gz" -C ~/dotnet
rm "dotnet-sdk-$DOTNET_VERSION-linux-x64.tar.gz"

PATH="~/dotnet:$PATH"

mv $SCRIPT_DIR/box.txt $SCRIPT_DIR/box.csproj
mv $SCRIPT_DIR/Program.txt $SCRIPT_DIR/Main.cs

dotnet restore $SCRIPT_DIR/box.csproj
dotnet publish -o $SCRIPT_DIR/output $SCRIPT_DIR/box.csproj
rm -rf $SCRIPT_DIR/output
rm $SCRIPT_DIR/setup.sh
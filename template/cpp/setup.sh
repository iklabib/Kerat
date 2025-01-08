#!/bin/bash
git clone https://github.com/google/googletest.git -b v1.15.2
mkdir googletest/build
cmake googletest -B googletest/build -DBUILD_GMOCK=OFF
make -C googletest/build
make -C googletest/build install
#!/bin/bash

# Copyright (c) 2024 Alibaba Group Holding Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

detect() {
    CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    CURRENT_ARCH=$(uname -m)
    if [ "${CURRENT_ARCH}" == "x86_64" ]; then
        CURRENT_ARCH="amd64"
    fi

    echo "Detected platform: ${CURRENT_OS} ${CURRENT_ARCH}"
}

download() {
    DOWNLOAD_URL="https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-${CURRENT_OS}-${CURRENT_ARCH}"
    EXECUTABLE="otel"

    echo "Downloading from $DOWNLOAD_URL"
    # curl and show progress
    curl -L -o "$EXECUTABLE" "$DOWNLOAD_URL"

    if [ $? -ne 0 ]; then
        echo "Failed to download $DOWNLOAD_URL"
        exit 1
    fi
}

install() {
    INSTALL_DIR="/usr/local/bin"
    if [ ! -f "$EXECUTABLE" ]; then
        echo "Executable $EXECUTABLE not found"
        exit 1
    fi
    echo "Installing $EXECUTABLE to $INSTALL_DIR"
    sudo mv "$EXECUTABLE" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$EXECUTABLE"

    echo "Installation completed. You can run it using: $INSTALL_DIR/$EXECUTABLE"
}

detect
download
install

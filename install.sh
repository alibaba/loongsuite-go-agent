# Copyright (c) 2024 Alibaba Group Holding Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash

set -x

CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)
if [ "${CURRENT_ARCH}" == "x86_64" ]; then
    CURRENT_ARCH="amd64"
fi

echo "Detected platform: ${CURRENT_OS} ${CURRENT_ARCH}"


DOWNLOAD_URL="https://github.com/alibaba/opentelemetry-go-auto-instrumentation/releases/download/v0.2.0/otelbuild-${CURRENT_OS}-${CURRENT_ARCH}"
EXECUTABLE="otelbuild"

echo "Downloading from $DOWNLOAD_URL"
wget -q --show-progress -O "$EXECUTABLE" "$DOWNLOAD_URL"

if [ $? -ne 0 ]; then
    echo "Failed to download $DOWNLOAD_URL"
    exit 1
fi

INSTALL_DIR="/usr/local/bin"
echo "Installing $EXECUTABLE to $INSTALL_DIR"
sudo mv "$EXECUTABLE" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$EXECUTABLE"

echo "Installation completed. You can run it using: $INSTALL_DIR/$EXECUTABLE"

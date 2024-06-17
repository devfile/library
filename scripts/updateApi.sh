#!/bin/bash

#
# Copyright Red Hat
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BLUE='\033[1;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

set -e

DIR=$(dirname $0)
CURRENT_DIR=$(pwd)
API_PKG="github.com/devfile/api/v2"
SCHEMA_URL_MASTER="https://raw.githubusercontent.com/devfile/api/main/schemas/latest/devfile.json"

# 2.0.0 devfile
SCHEMA_URL_200="https://raw.githubusercontent.com/devfile/api/2.0.x/schemas/latest/devfile.json"
PACKAGE_VERSION_200="version200"
JSON_SCHEMA_200="JsonSchema200"
FILE_PATH_200="$DIR/../pkg/devfile/parser/data/v2/2.0.0/devfileJsonSchema200.go"

# 2.1.0 devfile
SCHEMA_URL_210="https://raw.githubusercontent.com/devfile/api/2.1.x/schemas/latest/devfile.json"
PACKAGE_VERSION_210="version210"
JSON_SCHEMA_210="JsonSchema210"
FILE_PATH_210="$DIR/../pkg/devfile/parser/data/v2/2.1.0/devfileJsonSchema210.go"

# 2.2.0 devfile
SCHEMA_URL_220="https://raw.githubusercontent.com/devfile/api/2.2.x/schemas/latest/devfile.json"
PACKAGE_VERSION_220="version220"
JSON_SCHEMA_220="JsonSchema220"
FILE_PATH_220="$DIR/../pkg/devfile/parser/data/v2/2.2.0/devfileJsonSchema220.go"

# 2.2.1 devfile
SCHEMA_URL_221="https://raw.githubusercontent.com/devfile/api/2.2.x/schemas/latest/devfile.json"
PACKAGE_VERSION_221="version221"
JSON_SCHEMA_221="JsonSchema221"
FILE_PATH_221="$DIR/../pkg/devfile/parser/data/v2/2.2.1/devfileJsonSchema221.go"

# 2.2.2 devfile
SCHEMA_URL_222="https://raw.githubusercontent.com/devfile/api/2.2.x/schemas/latest/devfile.json"
PACKAGE_VERSION_222="version222"
JSON_SCHEMA_222="JsonSchema222"
FILE_PATH_222="$DIR/../pkg/devfile/parser/data/v2/2.2.2/devfileJsonSchema222.go"

# 2.3.0 devfile
PACKAGE_VERSION_230="version230"
JSON_SCHEMA_230="JsonSchema230"
FILE_PATH_230="$DIR/../pkg/devfile/parser/data/v2/2.3.0/devfileJsonSchema230.go"

onError() {
  cd "${CURRENT_DIR}"
}
trap 'onError' ERR


echo -e "${GREEN}Updating devfile/api in go.mod${NC}"
go get "${API_PKG}@main"

echo -e "${GREEN}Get latest schema${NC}"

case "${1}" in
   "2.0.0")
     SCHEMA_URL=${SCHEMA_URL_200}
     PACKAGE_VERSION=${PACKAGE_VERSION_200}
     JSON_SCHEMA=${JSON_SCHEMA_200}
     FILE_PATH=${FILE_PATH_200}
   ;;
    "2.1.0")
     SCHEMA_URL=${SCHEMA_URL_210}
     PACKAGE_VERSION=${PACKAGE_VERSION_210}
     JSON_SCHEMA=${JSON_SCHEMA_210}
     FILE_PATH=${FILE_PATH_210}
   ;;
   "2.2.0")
     SCHEMA_URL=${SCHEMA_URL_220}
     PACKAGE_VERSION=${PACKAGE_VERSION_220}
     JSON_SCHEMA=${JSON_SCHEMA_220}
     FILE_PATH=${FILE_PATH_220}
   ;;
   "2.2.1")
     SCHEMA_URL=${SCHEMA_URL_221}
     PACKAGE_VERSION=${PACKAGE_VERSION_221}
     JSON_SCHEMA=${JSON_SCHEMA_221}
     FILE_PATH=${FILE_PATH_221}
   ;;
   "2.2.2")
     SCHEMA_URL=${SCHEMA_URL_222}
     PACKAGE_VERSION=${PACKAGE_VERSION_222}
     JSON_SCHEMA=${JSON_SCHEMA_222}
     FILE_PATH=${FILE_PATH_222}
   ;;
   *)
     # default
     SCHEMA_URL=${SCHEMA_URL_MASTER}
     PACKAGE_VERSION=${PACKAGE_VERSION_230}
     JSON_SCHEMA=${JSON_SCHEMA_230}
     FILE_PATH=${FILE_PATH_230}
   ;;
esac

schema=$(curl -L "${SCHEMA_URL}")

#replace all ` with ' and write to schema file
echo -e "${GREEN}Write to go file${NC}"
go build $DIR/../*.go
./main updateSchema "${schema}" "${SCHEMA_URL}" "${PACKAGE_VERSION}" "${JSON_SCHEMA}" "${FILE_PATH}"


#!/bin/bash

# This script uses github_changelog_generator to generate a changelog and requires:
#
# 1. set an env GITHUB_TOKEN for the Github token
# 2. previous release as an arg, to generate changelog since the mentioned release
# 3. github_changelog_generator be installed where the script is being executed
#
# A CHANGELOG.md is generated and it's contents can be copy-pasted on the Github release

#TODO: Since issue tracking happens in devfile/api, github_changelog_generator cannot
# detect the issues from a different repository. Need to check if this is achievable.

BLUE='\033[1;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

# Ensure the github token is set
if [ -z "$GITHUB_TOKEN" ]
then
  echo -e "${RED}GITHUB_TOKEN env variable is empty..\nGet your GitHub token from https://github.com/settings/tokens and export GITHUB_TOKEN=<token>${NC}"
  exit 1
fi

# Ensure there is a release version passed in
if [ -z "$1" ]
then
  echo -e "${RED}The last release version needs to be provided. Changelog will be generated since that release..${NC}"
  echo -e "${RED}Example: ./changelog-script.sh v1.0.0-alpha.2 will generate a changelog for all the changes since release v1.0.0-alpha.2${NC}"
  exit 1
fi

# Ensure github_changelog_generator is installed
if ! command -v github_changelog_generator &> /dev/null
then
    echo -e "${RED}The command github_changelog_generator could not be found, please install the command to generate a changelog${NC}"
    exit 1
fi


github_changelog_generator \
-u devfile \
-p library \
-t $GITHUB_TOKEN \
--since-tag $1 \

RESULT=$?

if [ $RESULT -eq 0 ]; then
  echo -e "${GREEN}Changelog since release $1 generated at $PWD/CHANGELOG.md${NC}"
else
  echo -e "${RED}Unable to generate changelog using github_changelog_generator${NC}"
  exit 1
fi

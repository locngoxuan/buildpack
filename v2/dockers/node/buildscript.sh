#!/bin/sh
yarn_version=$(yarn --version)
echo "yarn version : ${yarn_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi
if [[ -z "${REVISION}" ]];then
	yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}
else
	sleep 1
fi
yarn install --cwd ${CWD}
yarn build --cwd ${CWD}
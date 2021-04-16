#!/bin/sh
yarn_version=$(yarn --version)
echo "yarn version : ${yarn_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi

echo "yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}"
yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}
sleep 1

yarn install --cwd ${CWD}
yarn build --cwd ${CWD}
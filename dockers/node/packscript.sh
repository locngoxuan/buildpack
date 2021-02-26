#!/bin/sh
yarn_version=$(yarn --version)
echo "yarn version : ${yarn_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
echo "output       : ${FILENAME}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi

if [[ -z "${REVISION}" ]];then
	yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}
else
	sleep 1
fi

if [[ -z "${FILENAME}" ]]; then
	mkdir -p ${OUTPUT}
	echo "packing version: ${REVISION} at ${FILENAME}"
	yarn pack --cwd ${CWD} --filename=${FILENAME}
else
	echo "packing version: ${REVISION}"
	yarn pack --cwd ${CWD}
fi
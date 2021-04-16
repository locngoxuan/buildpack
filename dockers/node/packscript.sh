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

echo "yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}"
yarn version --new-version ${REVISION} --no-git-tag-version --cwd ${CWD}
sleep 1

if [[ -z "${FILENAME}" ]]; then
    echo "packing version: ${REVISION}"
	yarn pack --cwd ${CWD}
else
    mkdir -p ${OUTPUT}
	echo "packing version: ${REVISION} at ${FILENAME}"
	yarn pack --cwd ${CWD} --filename=${FILENAME}
fi
#!/bin/sh
npm_version=$(npm --version)
echo "npm version : ${npm_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
echo "output       : ${FILENAME}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi

if [[ -z "${REVISION}" ]];then
	npm version ${REVISION} --git-tag-version=false --prefix ${CWD}
else
	sleep 1
fi

if [[ -z "${FILENAME}" ]]; then
	mkdir -p ${OUTPUT}
	echo "packing version: ${REVISION} at ${FILENAME}"
	npm pack ${FILENAME} --prefix ${CWD}
else
	echo "packing version: ${REVISION}"
	npm pack --prefix ${CWD}
fi
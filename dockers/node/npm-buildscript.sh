#!/bin/sh
npm_version=$(npm --version)
echo "npm version : ${npm_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi
if [[ -z "${REVISION}" ]];then
	npm version ${REVISION} --git-tag-version=false --prefix ${CWD}
else
	sleep 1
fi
npm install --prefix ${CWD}
npm run-script build --prefix ${CWD}
#!/bin/sh
npm_version=$(npm --version)
echo "npm version : ${npm_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi

echo "npm version ${REVISION} --git-tag-version=false --allow-same-version --prefix ${CWD}"
npm version ${REVISION} --git-tag-version=false --allow-same-version --prefix ${CWD}

sleep 1

npm install --prefix ${CWD}
npm run-script build --prefix ${CWD}
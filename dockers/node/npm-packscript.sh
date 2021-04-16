#!/bin/sh
npm_version=$(npm --version)
echo "npm version  : ${npm_version}"
echo "build version: ${REVISION}"
echo "working dir  : ${CWD}"
echo "output       : ${OUTPUT}"
echo "filename     : ${FILENAME}"

if [[ -z "${CWD}" ]];then
	echo "missing working directory"
	exit 1
fi

if [[ -z "${FILENAME}" ]];then
	echo "missing filename"
	exit 1
fi

echo "npm version ${REVISION} --git-tag-version=false --allow-same-version --prefix ${CWD}"
npm version ${REVISION} --git-tag-version=false --allow-same-version --prefix ${CWD}

sleep 1

#create output directory
mkdir -p ${OUTPUT}

#packing
echo "packing version: ${REVISION}"
npm pack --prefix ${CWD}

#move package to output directory
mv ${FILENAME} ${OUTPUT}/${FILENAME}
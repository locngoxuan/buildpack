#!/bin/sh
yarn --version

echo ${REVISION}
if [[ -z "${REVISION}" ]];then
	yarn version --new-version ${REVISION} --no-git-tag-version
else
	sleep 1
fi

if [[ -z "${FILENAME}" ]]; then
	echo "packing version: ${REVISION} at ${FILENAME}"
	yarn pack --filename=${FILENAME}
else
	echo "packing version: ${REVISION}"
	yarn pack
fi
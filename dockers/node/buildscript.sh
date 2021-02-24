#!/bin/sh
yarn --version
echo ${REVISION}
if [[ -z "${REVISION}" ]];then
	yarn version --new-version ${REVISION} --no-git-tag-version
else
	sleep 1
fi
yarn install
yarn build
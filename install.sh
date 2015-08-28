#!/usr/bin/env bash

APPNAME="app-twitter"
VERSION="v0.1.0"
LOCATION="apps"
FILENAME=${APPNAME}.tar.gz

echo "This script will download, install and run ${APPNAME}"
#sudo with-rw bash
cd /data/sphere/user-autostart/${LOCATION}
eval wget https://github.com/lindsaymarkward/${APPNAME}/releases/download/${VERSION}/${FILENAME}
mkdir ${APPNAME}
tar -xf ${FILENAME} -C ${APPNAME}
rm ${FILENAME}
nservice -q ${APPNAME} start
echo "Done... ${APPNAME} installed - hopefully :-)"
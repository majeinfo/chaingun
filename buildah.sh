#!/bin/bash
#
# Script for image building using buildah

ctr=$(buildah from debian:buster)
buildah run $ctr /bin/sh -c 'apt-get clean && apt-get update -y'
buildah run $ctr /bin/sh -c 'apt-get install -y git locales'
buildah run $ctr /bin/sh -c "sed -i '/^#.* fr_FR.UTF-8.* /s/^#//' /etc/locale.gen && locale-gen"

buildah run $ctr /bin/sh -c 'mkdir /scripts /output /data /appli' 
buildah run $ctr /bin/sh -c 'cd /appli && \
	git clone -b master https://github.com/majeinfo/chaingun.git && \
	cd chaingun && \
	rm -rf .git .gitignore Dockerfile start.sh player/src'

buildah config --workingdir /appli/chaingun $ctr
buildah run $ctr /bin/sh -c 'ln -s /data /appli/chaingun/manager/server/static/data'

buildah copy $ctr player/bin/player player/bin
buildah copy $ctr start.sh /

buildah config --env LANG=fr_FR.UTF-8 --env VERBOSE="" $ctr
buildah config -v /scripts -v /output -v /data $ctr
buildah config --port 8000/tcp $ctr

buildah config --entrypoint "/start.sh" $ctr

buildah commit $ctr newchaingun

# EOF

# BUILDER Image
FROM debian:buster as builder
RUN apt-get update -y
RUN apt-get install -y git golang

RUN mkdir /appli && cd /appli && git clone https://github.com/majeinfo/chaingun.git
WORKDIR /appli/chaingun

RUN export GOPATH=/appli/chaingun/player && \
	cd /appli/chaingun/player/src/github.com && \
	rm -rf gorilla sirupsen tobyhede && \
	cd /appli/chaingun/player/src/github.com/majeinfo/chaingun/player && \
	go get -d && \
	cd /appli/chaingun/player && \
	go install github.com/majeinfo/chaingun/player


# CHAINGUN Image
FROM debian:buster
LABEL maintainer "jd@maje.biz"

RUN apt-get update -y
#RUN apt-get install -y git golang python3 python3-pip locales
RUN apt-get install -y git python3 python3-pip locales
RUN sed -i '/^#.* fr_FR.UTF-8.* /s/^#//' /etc/locale.gen && locale-gen

RUN mkdir /scripts /output /data /appli && \
	cd /appli && \
	git clone https://github.com/majeinfo/chaingun.git && \
	rm -rf Dockerfile start.sh player/src

WORKDIR /appli/chaingun
RUN pip3 install -r requirements.txt
#RUN export GOPATH=/appli/chaingun/player && \
#	cd /appli/chaingun/player/src/github.com && \
#	rm -rf gorilla sirupsen tobyhede && \
#	cd /appli/chaingun/player/src/github.com/majeinfo/chaingun/player && \
#	go get -d && \
#	cd /appli/chaingun/player && \
#	go install github.com/majeinfo/chaingun/player

RUN ln -s /usr/bin/python3 /usr/bin/python
RUN ln -s /data /appli/chaingun/manager/server/static/data
COPY --from=builder /appli/chaingun/player/bin/player player/bin/
ADD start.sh /

ENV LANG fr_FR.UTF-8
ENV VERBOSE ""
VOLUME /scripts
VOLUME /output
VOLUME /data
EXPOSE 8000

ENTRYPOINT [ "/start.sh" ]

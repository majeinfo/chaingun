FROM debian:buster
LABEL maintainer "jd@maje.biz"

RUN apt-get update -y
RUN apt-get install -y git golang python3 python3-pip locales
RUN sed -i '/^#.* fr_FR.UTF-8.* /s/^#//' /etc/locale.gen && \
	locale-gen
RUN mkdir /appli && cd /appli && git clone https://github.com/majeinfo/chaingun.git
WORKDIR /appli/chaingun
RUN pip3 install -r requirements.txt
RUN mkdir /scripts && mkdir /output
RUN export GOPATH=/appli/chaingun/player && \
	cd /appli/chaingun/player/src/github.com && \
	rm -rf gorilla sirupsen tobyhede && \
	cd /appli/chaingun/player/src/github.com/majeinfo/chaingun/player && \
	go get -d && \
	cd /appli/chaingun/player && \
	go install github.com/majeinfo/chaingun/player
ADD start.sh /
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN mkdir /data && ln -s /data /appli/chaingun/manager/server/static/data

ENV LANG fr_FR.UTF-8
ENV VERBOSE ""
VOLUME /scripts
VOLUME /output
VOLUME /appli/chaingun/manager/server/static/data
EXPOSE 8000

ENTRYPOINT [ "/start.sh" ]

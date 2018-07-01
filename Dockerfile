FROM debian:buster
LABEL maintainer "jd@maje.biz"

RUN apt-get update -y
RUN apt-get install -y git golang python3 python3-pip
RUN mkdir /appli && cd /appli && git clone https://github.com/majeinfo/chaingun.git
WORKDIR /appli/chaingun
RUN pip3 install -r requirements.txt
RUN mkdir /scripts && mkdir /output
RUN export GOPATH=/appli/chaingun/player && \
	cd /appli/chaingun/player/src/github.com/majeinfo/chaingun/player && \
	go get -d && \
	cd /appli/chaingun/player && \
	go install github.com/majeinfo/chaingun/player

ENV VERBOSE ""
VOLUME /scripts
VOLUME /output
EXPOSE 8000

ENTRYPOINT [ "/start.sh" ]

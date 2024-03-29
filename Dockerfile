# BUILDER Image
FROM golang:1.16 as builder

RUN mkdir /appli && cd /appli && git clone -b master https://github.com/majeinfo/chaingun.git
WORKDIR /appli/chaingun

RUN export GOPATH=/appli/chaingun/player && \
	cd player/src && \
	go install ./player


# CHAINGUN Image
FROM debian:buster
LABEL maintainer "jd@maje.biz"

RUN apt-get clean && apt-get update -y
RUN apt-get install -y git locales
RUN sed -i '/^#.* fr_FR.UTF-8.* /s/^#//' /etc/locale.gen && locale-gen

RUN mkdir /scripts /output /data /appli && \
	cd /appli && \
	git clone -b master https://github.com/majeinfo/chaingun.git && \
	cd chaingun && \
	rm -rf .git .gitignore Dockerfile start.sh samples tests player/src

WORKDIR /appli/chaingun

#RUN ln -s /data /appli/chaingun/manager/server/static/data
COPY --from=builder /appli/chaingun/player/bin/player player/bin/
ADD start.sh /

ENV LANG fr_FR.UTF-8
ENV VERBOSE ""
VOLUME /scripts
VOLUME /output
VOLUME /data
EXPOSE 8000

ENTRYPOINT [ "/start.sh" ]

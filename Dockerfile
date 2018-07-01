FROM debian:8.11
LABEL maintainer "jd@maje.biz"

ARG target_dir=/appli
#RUN yum update -y
#RUN yum install -y git golang
#RUN yum clean all
RUN apt-get update -y
RUN apt-get install -y git golang
RUN mkdir ${target_dir} && cd ${target_dir} && git clone https://github.com/majeinfo/chaingun.git
EXPOSE 8000
WORKDIR ${target_dir}/chaingun
CMD [ "/usr/local/bin/python", "manage.py", "runserver", "0.0.0.0:8000" ]

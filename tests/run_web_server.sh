#!/bin/bash

docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache


#!/bin/bash

rm -r ~/.gnxi/*
docker stop gnoi_cert gnoi_os gnoi_reset
docker container rm gnoi_cert gnoi_os gnoi_reset
docker images -a | grep "gnoi_*" | awk '{print $3}' | xargs docker rmi

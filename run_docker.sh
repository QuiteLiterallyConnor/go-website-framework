#! /bin/bash

    docker rm resume_server
    go build main.go
    docker build -t resume_server .
    docker run -it --name resume_server resume_server
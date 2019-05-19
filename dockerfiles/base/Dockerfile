FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y \
	ca-certificates \
	build-essential \
	curl \
	make \
	vim

WORKDIR /tmp
RUN curl -sL -o go.tgz 'https://dl.google.com/go/go1.12.5.linux-amd64.tar.gz' && \
	echo "aea86e3c73495f205929cfebba0d63f1382c8ac59be081b6351681415f4063cf  go.tgz" | shasum -a 256 -c - && \
	tar -C /usr/local -xzf go.tgz && \
	rm go.tgz
ENV PATH /usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/bitsgofer/example

RUN apt-get install -y \
	netcat

# FROM alpine:3.9

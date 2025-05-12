FROM golang:1.23 AS build

WORKDIR /app

COPY . .

RUN go mod download

ENV GOOS=linux
RUN mkdir -p ./bin && go build -o ./bin/main main.go

FROM public.ecr.aws/lambda/provided:latest

WORKDIR /app

RUN dnf update -y
RUN dnf install -y zip tar findutils

# note: file will be invoked by crontab on host machine
# 0 4 * * * doker exec <name> /app/clean-stale.sh
RUN echo -e '#!/bin/sh\nfind /app/tmp -type d -name "nodelayer-*" -mtime +1 -exec rm -rf "{}" \;' > ./clean-stale.sh && \
	chmod 777 ./clean-stale.sh

RUN mkdir -p ./tmp && chmod 777 ./tmp
RUN mkdir -p ./bin && chmod 777 ./bin

COPY --from=build /app/bin/main /app/bin/main

EXPOSE 1923

ENV TMPDIR=/app/tmp
ENTRYPOINT ["/app/bin/main"]

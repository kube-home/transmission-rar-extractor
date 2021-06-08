FROM golang:alpine as build
WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o unzipper-amd64 .
RUN GOOS=linux GOARCH=arm64 go build -o unzipper-arm64 .

FROM alpine as app
RUN apk update && apk add unrar
COPY --from=build --chown=1000:1000 /app/unzipper-* /usr/bin/
USER 1000:1000
ENV GODEBUG asyncpreemptoff=1
CMD ["unzipper-amd64"]
FROM golang:1.22-alpine as golangchatapp

#RUN echo https://mirror.arvancloud.ir/alpine/v3.17/main > /etc/apk/repositories
#RUN echo https://mirror.arvancloud.ir/alpine/v3.17/community >> /etc/apk/repositories

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go

EXPOSE 8080
CMD ./main



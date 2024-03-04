FROM git.oteacher.org:5001/oteacher/devops/image-hub/golang:1.22 as golangchatapp

RUN echo https://mirror.arvancloud.ir/alpine/v3.17/main > /etc/apk/repositories
RUN echo https://mirror.arvancloud.ir/alpine/v3.17/community >> /etc/apk/repositories

ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go



FROM git.oteacher.org:5001/oteacher/devops/image-hub/golang:1.22 as chat
WORKDIR /app
COPY --from=GolangChatApp /app/main .
EXPOSE 8080
CMD ./main



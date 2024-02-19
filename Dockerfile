FROM golang:1.22-alpine as golangchatapp
ENV GO111MODULE=on
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go



FROM golang:1.22-alpine as chat
WORKDIR /app
COPY --from=GolangChatApp /app/main .
EXPOSE 8080
CMD ./main



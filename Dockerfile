FROM golang:1.16

WORKDIR /lambda

COPY . .

EXPOSE 8686

RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

# CMD [ "go", "build", "./cmd/lambda.go"]
ENTRYPOINT CompileDaemon -log-prefix=true -build="go build ./cmd/lambda.go" -command="./lambda"
FROM golang:1.16

WORKDIR /lambda

COPY . .

EXPOSE 8686

RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

#CMD [ "go", "build", "."]
ENTRYPOINT CompileDaemon -log-prefix=false -build="go build ." -command="./dgraph-lambda-go run"
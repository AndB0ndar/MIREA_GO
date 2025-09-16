mkdir -p ~/helloapi/cmd/server
cd ~/helloapi

go mod init example.com/helloapi

#vim ~/helloapi/cmd/server/main.go

go get github.com/google/uuid@latest
go mod tidy

curl http://localhost:8081/hello
curl http://localhost:8081/user

go build -o helloapi ./cmd/server
./helloapi

go fmt ./...
go vet ./...

APP_PORT=9090 go run ./cmd/server
curl http://localhost:9090/hello
curl http://localhost:9090/user


CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o tchain-arm .

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o tchain-amd .

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o tchain-windows .
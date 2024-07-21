CLIENT=./cmd/client/main.go

build_win_client: 
	 set GOOS=windows
	 set GOARCH=amd64
	 go build -o ./bin/gophkeeper.exe ${CLIENT}
build_mac_client: 
	 set GOOS=darwin
	 set GOARCH=arm64
	 go build -o ./bin/gophkeeper_osx ${CLIENT}
build_linux_client: 
	 set GOOS=linux 
	 set GOARCH=amd64 
	 go build -o ./bin/gophkeeper_linux ${CLIENT}
build_certs:
	cd cert; ./gen.sh; cd ..
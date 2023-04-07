VERSION=v0.0.1

.PHONY: bin
bin: bin/kubectl-statefulreplace_darwin_x86_64 bin/kubectl-statefulreplace_darwin_arm64 bin/kubectl-statefulreplace_linux_x86_64 bin/kubectl-statefulreplace_linux_arm bin/kubectl-statefulreplace_linux_arm64 bin/kubectl-statefulreplace_windows_x86_64.exe
bin: bin/kubectl-statefulreplace

.PHONY: install
install: bin/kubectl-statefulreplace
	cp bin/kubectl-statefulreplace /usr/local/bin/kubectl-statefulreplace
	chmod +x /usr/local/bin/kubectl-statefulreplace

bin/kubectl-statefulreplace:
	mkdir -p bin
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace > bin/kubectl-statefulreplace.sha512

bin/kubectl-statefulreplace_darwin_x86_64:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_darwin_x86_64 cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_darwin_x86_64 > bin/kubectl-statefulreplace_darwin_x86_64.sha512

bin/kubectl-statefulreplace_darwin_arm64:
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_darwin_arm64 cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_darwin_arm64 > bin/kubectl-statefulreplace_darwin_arm64.sha512

bin/kubectl-statefulreplace_linux_x86_64:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_linux_x86_64 cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_linux_x86_64 > bin/kubectl-statefulreplace_linux_x86_64.sha512

bin/kubectl-statefulreplace_linux_arm:
	mkdir -p bin
	GOOS=linux GOARCH=arm go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_linux_arm cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_linux_arm > bin/kubectl-statefulreplace_linux_arm.sha512

bin/kubectl-statefulreplace_linux_arm64:
	mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_linux_arm64 cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_linux_arm64 > bin/kubectl-statefulreplace_linux_arm64.sha512

bin/kubectl-statefulreplace_windows_x86_64.exe:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/kubectl-statefulreplace_windows_x86_64.exe cmd/statefulreplace/*.go
	openssl sha512 bin/kubectl-statefulreplace_windows_x86_64.exe > bin/kubectl-statefulreplace_windows_x86_64.exe.sha512
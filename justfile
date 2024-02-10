alias b := build
alias r := run

build:
  set -e
  go mod download
  go build -o ./main cmd/main.go
fmt:
  find . -name "*.go" -exec go fmt {} \;
run: build
  ./main


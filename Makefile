.PHONY: test testdel example install release
testdel:
	go test ./... -coverprofile=cover.out -v
	go tool cover -func=cover.out
	rm -f cover.out
test:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out
example:
	cd example && rm -Rf data && go run *.go && cd -
install:
	go install github.com/fobilow/gitdb/v2/cmd/gitdb
release:
	go install github.com/fobilow/gitdb/v2/cmd/gitdb
	go generate

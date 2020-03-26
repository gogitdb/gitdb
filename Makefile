.PHONY: test
testdel:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out
	rm -f cover.out
test:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out
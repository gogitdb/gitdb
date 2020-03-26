.PHONY: test
test:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out
	rm -f cover.out
.PHONY: example
example:
	cd example && rm -Rf data && go run main.go && cd -
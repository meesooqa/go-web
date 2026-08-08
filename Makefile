test:
	go clean -testcache
	go test ./...

upd:
	go get -u ./...
	go mod tidy
	go mod vendor

.PHONY: test upd
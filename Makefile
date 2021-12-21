
.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0

.PHONY: lint
lint:
	golangci-lint run

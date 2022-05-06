# Developer helpers

.PHONY: test
test:
	go test ./...

.PHONY: install-test
install-test:
	go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps

.PHONY: install-dev
install-dev:
	go install mvdan.cc/gofumpt@latest

.PHONY: format
format:
	go mod tidy
	gofumpt -l -w ./

# Run static code analysis
.PHONY: lint
lint:
	golangci-lint run

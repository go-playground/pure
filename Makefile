GOCMD=GO111MODULE=on go

linters-install:
	@golangci-lint --version >/dev/null 2>&1 || { \
		echo "installing linting tools..."; \
		$(GOCMD) get github.com/golangci/golangci-lint/cmd/golangci-lint; \
	}

lint: linters-install
	golangci-lint run

test:
	$(GOCMD) test -cover -race ./...

bench:
	$(GOCMD) test -bench=. -benchmem ./...

.PHONY: linters-install lint test bench
GOCMD=GO111MODULE=on go

lint:
	golint -set_exit_status ./...

test:
	$(GOCMD) test -cover -race ./...

bench:
	$(GOCMD) test -bench=. -benchmem ./...

.PHONY: lint test bench
.PHONY: init fmt vet lint precommit

init:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please provide a name. Example: make init NAME=github.com/dev-au/CodeStream"; \
		exit 1; \
	fi
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File ./scripts/init.ps1 -NewName $(NAME)
else
	@chmod +x ./scripts/init.sh
	@./scripts/init.sh $(NAME)
endif

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

lint:
	golangci-lint run --fix

arch-check:
	./scripts/arch-check.sh

test:
	go test -v ./...

precommit: tidy lint arch-check test

run:
	go run ./cmd
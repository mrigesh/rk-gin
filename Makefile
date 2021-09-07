.PHONY: all
all: gomod test lint readme fmt

.PHONY: lint
lint:
	@echo "running golangci-lint..."
	@golangci-lint run 2>&1

.PHONY: test
test:
	@echo "running go test..."
	@go test ./... -coverprofile coverage.txt 2>&1

.PHONY: fmt
fmt:
	@echo "format go project..."
	@gofmt -s -w . 2>&1

.PHONY: readme
readme:
	@echo "running doctoc..."
	@doctoc . 2>&1

.PHONY: swag
swag:
	@echo "running swag..."
	@swag init --generalInfo boot/common_service_entry.go --output boot/assets/sw/config --parseDependency 2 --propertyStrategy camelcase

.PHONY: pkger
pkger:
	@echo "running pkger..."
	@pkger -o boot

.PHONY: gomod
gomod:
	@echo "running go mod tidy..."
	@go mod tidy
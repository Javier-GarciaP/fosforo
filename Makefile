BIN     := bin/fos
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
LDFLAGS := -X github.com/Javier-GarciaP/fosforo/internal/version.Version=$(VERSION) \
           -X github.com/Javier-GarciaP/fosforo/internal/version.Commit=$(COMMIT)

.PHONY: build test vet fmt run

build: ## Compila el binario en bin/fos
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/fos

test: ## Ejecuta todas las pruebas
	go test -race ./...

vet: ## Análisis estático y comprobación de formato
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "archivos sin formatear:"; gofmt -l .; exit 1)

fmt: ## Da formato a todo el código
	gofmt -w .

run: build ## Compila y ejecuta (make run ARGS="README.md")
	./$(BIN) $(ARGS)

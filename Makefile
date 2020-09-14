build:
ifeq ($(OS),Windows_NT)
	go build -o build/swap-backend.exe cmd/swap-backend/main.go
else
	go build -o build/swap-backend cmd/swap-backend/main.go
endif

install:
ifeq ($(OS),Windows_NT)
	go install cmd/swap-backend/main.go
else
	go install cmd/swap-backend/main.go
endif

.PHONY: build install

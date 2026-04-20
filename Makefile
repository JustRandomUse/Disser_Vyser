.PHONY: help install dev build clean

help:
	@echo "Available commands:"
	@echo "  make install    - Install all dependencies"
	@echo "  make dev        - Run development servers"
	@echo "  make build      - Build for production"
	@echo "  make clean      - Clean build artifacts"

install:
	@echo "Installing backend dependencies..."
	cd back && go mod download
	@echo "Installing frontend dependencies..."
	cd front && npm install

dev-backend:
	@echo "Starting backend server..."
	cd back && go run cmd/server/main.go

dev-frontend:
	@echo "Starting frontend dev server..."
	cd front && npm run dev

build-frontend:
	@echo "Building frontend..."
	cd front && npm run build

build-backend:
	@echo "Building backend..."
	cd back && go build -o server cmd/server/main.go

build: build-frontend build-backend
	@echo "Build complete!"

clean:
	@echo "Cleaning..."
	rm -rf front/dist
	rm -f back/server
	@echo "Clean complete!"

run-prod: build
	@echo "Starting production server..."
	cd back && ./server

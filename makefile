.PHONY: build deploy pull

include .env
export

build:
	@echo "⚙️ Starting local Go build..."
	GOOS=linux GOARCH=amd64 go build -o ./build/scraper ./cmd/scraper

deploy: build
	@echo "🚀 Deploying..."
	./deploy.sh

pull:
	scp -P $(SSH_PORT) -r $(SSH_USER)@$(SSH_HOST):$(REMOTE_DEPLOY_DIR)/database.db ./database.db
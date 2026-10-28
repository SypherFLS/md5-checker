DB_URL := $(DATABASE_URL)
.PHONY: migrate, bups, bup, down, callv, lint, prep, gpush

migrate:
	cd backend && goose -dir ./migrations postgres "$(DB_URL)" up

bups: # пересобрать бекенд
	cd backend && docker build -t thread-service . && docker-compose up -d --build 'backend'

bup: # собрать весь проект
	cd backend && docker build -t thread-service . && docker-compose up -d --build 

down: 
	cd backend && docker-compose down -v

callv:
	cd backend && go-callvis ./cmd/main.go

lint:
	cd backend && golangci-lint run

prep:
	cd backend && go fmt ./...
 	
gpush:
	git add ${dir} backend/go.mod backend/go.sum
	git commit -m "${msg}"
	git push

.PHONY: run test docker-up docker-down migrate-up migrate-down swagger clean

# Serverni ishga tushirish
run:
	go run cmd/api/main.go

# Barcha testlarni ishga tushirish
test:
	go test -v -race ./tests/...

# Docker PostgreSQL ishga tushirish
docker-up:
	docker compose up -d

# Docker PostgreSQL to'xtatish
docker-down:
	docker compose down

# Migratsiyalarni amalga oshirish
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

# Migratsiyalarni qaytarish
migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down

# Swagger hujjatlarini generatsiya qilish
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Tozalash
clean:
	go clean

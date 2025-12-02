# Makefile for rainy

run:
	go run cmd/main.go

migrate-up:
	go run db/migrate.go $(DB) up

migrate-down:
	go run db/migrate.go $(DB) down

migrate-reset:
	go run db/migrate.go $(DB) reset

gen-key-rsa256:
	openssl genpkey -algorithm RSA -out certs/private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in certs/private.pem -out certs/public.pem
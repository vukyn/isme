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
	mkdir -p certs
	openssl genpkey -algorithm RSA -out certs/private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in certs/private.pem -out certs/public.pem

init-ui:
	npm create vite@latest ui
	cd ui && npm i @chakra-ui/react @emotion/react
	cd ui && npx @chakra-ui/cli snippet add

run-ui:
	cd ui && npm run dev

build-ui:	
	cd ui && npm run build
	rm -rf ./ui
	mv ./ui/dist ./ui
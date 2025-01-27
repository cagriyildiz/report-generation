include .env

db_run:
	docker run --name postgres-db-17 -p ${DB_PORT}:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASS} -d postgres:17-alpine

db_create:
	docker exec -it postgres-17 createdb --username=${DB_NAME} --owner=${DB_NAME} ${DB_NAME}

db_drop:
	 docker exec -it postgres-17 dropdb ${DB_NAME}

db_migrate_create:
	migrate create -ext sql -dir db/migration -seq $(name)

migrate_up:
	migrate -path db/migration -database ${DB_URL} -verbose up

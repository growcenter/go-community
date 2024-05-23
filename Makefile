docker-start:
	# Create database migration
	docker-compose up -d

docker-stop:
	# Delete postgre database server
	docker-compose down

run_api: tidy
	export ENV="DEV" && go run ./cmd/api/main.go

tidy:
	go mod tidy
	go mod download

migration:
	# Create database migration
	migrate create -ext sql -dir tests/integration/db/migrations/ -seq ${name}

database_up:
	# Create database
	docker exec -it community createdb --username=postgres --owner=postgres community_db

database_down:
	# Create database
	docker exec -it community dropdb --username=postgres community_db

migration_up:
	# Create table for community_db
	migrate -path tests/integration/db/migrations/ -database "postgres://postgres:<<password>>@<<host>>:<<port>>/postgres?sslmode=disable" up

migration_down:
	# Create table for community_db
	migrate -path tests/integration/db/migrations/ -database "postgres://postgres:<<password>>@<<host>>:<<port>>/postgres?sslmode=disable" down

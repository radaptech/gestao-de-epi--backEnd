migration:
	@migrate create -ext  sql -dir database/migrate -seq $(filter-out $@, $(MAKECMDGOALS)) 

migrate-up:
	@go run main.go Up
migrate-down:
	@go run main.go Down


backup-banco:
	@go run . backup-banco
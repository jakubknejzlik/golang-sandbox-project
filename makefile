migrate-db:
	DATABASE_URL=sqlite3://test.db go run *.go migrate

start:
	DATABASE_URL=sqlite3://test.db PORT=8080 go run *.go start

test:
	go test ./src

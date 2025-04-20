start:
	docker compose up --build

test:
	mkdir -p coverage
	go test -v $$(go list ./... | grep -Ev '/(mocks|docs|cmd|db|config|internal/app)') -coverprofile=coverage/cover.out

coverage: test
	go tool cover -html=coverage/cover.out -o coverage/cover.html
	go tool cover -func=coverage/cover.out | grep total:

clean-coverage:
	rm -rf coverage/
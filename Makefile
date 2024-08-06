generate: generate-css generate-go

generate-go:
	go generate -x ./...

generate-css:
	pnpm exec tailwindcss -i input.css -o public/styles.css

deps:
	go mod tidy -v

dev:
	go run github.com/air-verse/air

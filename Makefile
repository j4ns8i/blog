generate: generate-css generate-go

generate-go:
	go generate -x ./...

generate-css:
	pnpm dlx tailwindcss -i input.css -o public/styles.css

build_wasm:
	GOOS=js GOARCH=wasm go build -C ./client/wasm/ -o ../static/wasm/main.wasm

ts_watch:
	cd client/ts && tsc --watch

templ_watch:
	cd client/app && templ generate --watch ./... --proxy=http://localhost:8080

tw_watch:
	cd client/app && npx tailwindcss -i ../static/css/input.css -o ../static/css/output.css --watch

app_watch:
	cd client/app && air

wasm_watch:
	cd client/wasm && air

FROM golang:1.25 AS build-stage
WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY /app /app

RUN go install github.com/a-h/templ/cmd/templ@latest && templ generate ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint

FROM tinygo/tinygo:0.32.0 AS build-wasm-stage
WORKDIR /wasm
USER root
COPY wasm .

RUN mkdir -p /static/wasm
RUN mkdir -p /static/js

RUN GOOS=js GOARCH=wasm tinygo build -o /static/wasm/main.wasm

# RUN mkdir -p /static/js && cp $(go env GOROOT)/misc/wasm/wasm_exec.js /static/js/wasm_exec.js
RUN cp /usr/local/tinygo/targets/wasm_exec.js /static/js/wasm_exec.js

FROM node:20 AS build-js-stage
WORKDIR /ts
COPY /ts /ts
RUN npm install
RUN npm run build:docker

WORKDIR /app
COPY /app /app
COPY /static/css/input.css /static/css/input.css
RUN npm install
RUN npm run tailwind:generate

FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /

COPY --from=build-stage /entrypoint /entrypoint
COPY --from=build-wasm-stage /static /static
COPY --from=build-js-stage /ts/dist /static/js
COPY --from=build-js-stage /static /static

ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]

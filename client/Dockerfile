FROM golang:1.22 AS build-stage
WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY /app /app

RUN go install github.com/a-h/templ/cmd/templ@latest && templ generate ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint

WORKDIR /wasm
COPY wasm .
RUN GOOS=js GOARCH=wasm go build -o /dist/static/wasm/main.wasm

RUN mkdir -p /dist/static/js && cp $(go env GOROOT)/misc/wasm/wasm_exec.js /dist/static/js/wasm_exec.js

FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /

COPY --from=build-stage /entrypoint /entrypoint
COPY /static /static

ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]

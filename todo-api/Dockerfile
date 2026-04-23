# syntax=docker/dockerfile:1.6

# ---------- build stage ----------
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build a fully static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/todo-api .

# ---------- runtime stage ----------
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/todo-api /app/todo-api

EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/app/todo-api"]

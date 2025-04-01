# build stage
FROM golang:alpine AS build
WORKDIR /app
# Copy only go mod files first
COPY go.mod go.sum ./
RUN go mod download
# Copy the source code
COPY . .
RUN go build -o output/hermes ./cmd/main.go

# run stage
FROM cloudflare/cloudflared:latest
COPY --from=build /app/output/hermes /bin/hermes
ENTRYPOINT ["hermes"]

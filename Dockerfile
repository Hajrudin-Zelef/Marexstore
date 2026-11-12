FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/store ./cmd/store && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/import ./import

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=build /out/store /app/store
COPY --from=build /out/import /app/import
COPY templates ./templates
COPY static ./static
COPY db ./db
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/store"]

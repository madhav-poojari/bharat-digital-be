# build stage
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG PACKAGE=./cmd/api    # change if your main is elsewhere, e.g. ./ or ./cmd/myapi
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /server $PACKAGE

# runtime stage
FROM scratch
COPY --from=build /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]

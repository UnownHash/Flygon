# Build image
FROM golang:1.20-alpine as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/flygon

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11 as runner
COPY --from=build /go/bin/flygon /flygon/
COPY /sql /flygon/sql
COPY /logs /flygon/logs

WORKDIR /flygon
CMD ["./flygon"]

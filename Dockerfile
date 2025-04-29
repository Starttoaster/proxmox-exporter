FROM golang:1 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/app /
CMD ["/app"]

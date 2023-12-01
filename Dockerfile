FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY  ./ ./

RUN go clean --modcache && go build -ldflags "-w" -mod=readonly -o /bin cmd/worker/app.go

FROM scratch

WORKDIR /
COPY --from=builder  /bin /bin

EXPOSE 8080

CMD ["bin/app"]
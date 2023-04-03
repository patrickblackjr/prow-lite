FROM golang:1.20.2-bullseye
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# RUN go build -v -o /usr/local/bin/app ./...
CMD ["go", "run", "."]

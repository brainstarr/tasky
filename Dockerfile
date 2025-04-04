# Building the binary of the App
FROM golang:1.19 AS build

WORKDIR /go/src/tasky
COPY . .
RUN go mod download

#create wizexercise.txt file

RUN echo "WIZ | TECHNICAL EXERCISE v3.0" > wizexercise.txt

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/src/tasky/tasky


FROM alpine:3.17.0 as release

WORKDIR /app
COPY --from=build  /go/src/tasky/tasky .
COPY --from=build  /go/src/tasky/assets ./assets

#copy wizexercise.txt file into release image
COPY --from=build /go/src/tasky/wizexercise.txt .

EXPOSE 8080
ENTRYPOINT ["/app/tasky"]

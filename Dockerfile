FROM golang:1.21.0-bookworm as build
WORKDIR /src
COPY . ./
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o app.bin .

FROM gruebel/upx:latest as upx
WORKDIR /src
COPY --from=build /src/app.bin ./
RUN upx -9 --lzma ./app.bin

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=upx /src/app.bin ./app
COPY --from=build /src/config.json ./
CMD [ "./app" ]
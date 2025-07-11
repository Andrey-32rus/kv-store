FROM tarantool/tarantool:latest AS tarantool

FROM golang:1.24.1 as builder
WORKDIR /app
COPY . .
RUN go build -o kvstore

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y libmsgpuck-dev curl

# Copy Tarantool binary
COPY --from=tarantool /usr/bin/tarantool /usr/bin/tarantool
# Copy Go app
COPY --from=builder /app/kvstore /kvstore

# Copy Tarantool init script
COPY init.lua /init.lua

EXPOSE 8080

# Entrypoint script: start Tarantool and Go app
CMD tarantool /init.lua & sleep 3 && ./kvstore

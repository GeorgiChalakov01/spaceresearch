FROM alpine:latest

RUN apk add --no-cache poppler-utils
WORKDIR /app
COPY src/app /app
ENTRYPOINT ["/app/app"]

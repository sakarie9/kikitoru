FROM alpine
COPY kikitoru /app/kikitoru

RUN apk add --no-cache gcompat tzdata

ENTRYPOINT ["/app/kikitoru"]

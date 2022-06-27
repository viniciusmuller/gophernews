FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .
RUN adduser -S -D -H -h /app appuser

# TODO: Add make migrations and data persistence work
# RUN apk add migrateo
# RUN migrate -database $POSTGRESQL_URL -path db/migrations up
USER appuser
CMD ["./main"]

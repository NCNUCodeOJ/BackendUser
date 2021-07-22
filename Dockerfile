# build stage
FROM golang:1.15 AS build-env
COPY . /src
RUN cd /src && go build -o app

# final stage
FROM alpine:3
WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
COPY --from=build-env /src/app /app
ENTRYPOINT ./app
HEALTHCHECK --timeout=5s CMD ./app ping
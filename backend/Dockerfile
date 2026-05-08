FROM golang:1.25-alpine AS builder
WORKDIR /src/backend
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
FROM python:3.12-slim AS runtime
WORKDIR /app

# fastembed
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates libgomp1 \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/api /app/api

# Python embedding dependencies
COPY embedding-service/requirements.txt /app/embedding-service/requirements.txt
RUN python -m pip install --no-cache-dir -r /app/embedding-service/requirements.txt
COPY embedding-service/app /app/embedding-service/app

#Runtime directories
RUN useradd --uid 10001 --create-home --shell /usr/sbin/nologin appuser \
  && mkdir -p /app/uploads /app/.cache \
  && chown -R appuser:appuser /app/uploads /app/.cache
EXPOSE 8080
ENV PORT=8080 \
  PYTHON_PATH=python3 \
  SCRIPT_PATH=/app/embedding-service/app/main.py \
  PYTHONUNBUFFERED=1 \
  XDG_CACHE_HOME=/app/.cache
USER 10001:10001
CMD ["/app/api"]

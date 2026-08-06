# syntax=docker/dockerfile:1
# Optimized Dockerfile with improved layer caching and BuildKit secrets
# Changes:
# 1. Better layer ordering for cache optimization
# 2. Use BuildKit --mount=type=secret for private repo auth (more secure)
# 3. Use BuildKit cache mounts for Go modules and build cache
# 4. Consolidated RUN commands where beneficial

# Stage 1: Builder
FROM golang:1.26.3-alpine3.23 AS builder

# Set workdir
WORKDIR /app

# Layer 1: Install system dependencies (rarely changes - highest cache)
RUN apk update --no-cache && apk add --no-cache git openssh

# Layer 2: Install Delve debugger (rarely changes)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest

# Layer 3: Copy go.mod and go.sum FIRST (changes when dependencies update)
COPY go.mod go.sum ./

# Layer 4: Download dependencies with secret mount for private repos
# Uses BuildKit secret mount - more secure than ARG
# Requires: docker build --secret id=github_token,src=/path/to/token
RUN --mount=type=secret,id=github_token \
    --mount=type=cache,target=/go/pkg/mod \
    git config --global url."https://$(cat /run/secrets/github_token)@github.com/".insteadOf "https://github.com/" && \
    export GOPRIVATE=github.com/paper-indonesia/* && \
    go mod download -x

# Layer 5: Copy source code (frequently changes - lowest cache)
COPY . .

# Layer 6: Build application with cache mounts
RUN --mount=type=secret,id=github_token \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    git config --global url."https://$(cat /run/secrets/github_token)@github.com/".insteadOf "https://github.com/" && \
    export GOPRIVATE=github.com/paper-indonesia/* && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /app/backend-portal .

# Layer 7: Build debug version
RUN --mount=type=secret,id=github_token \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    git config --global url."https://$(cat /run/secrets/github_token)@github.com/".insteadOf "https://github.com/" && \
    export GOPRIVATE=github.com/paper-indonesia/* && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -a -installsuffix cgo -o /app/backend-portal_debug .

# Stage 2: Runtime (Final Image)
FROM alpine:3.23

# Layer 1: Create non-root user (rarely changes)
RUN addgroup -S user && adduser -S user -G user

WORKDIR /app

# Layer 2: Set ownership (rarely changes)
RUN chown -R user:user /app

# Layer 3: Install runtime dependencies (rarely changes)
RUN apk add --no-cache ca-certificates tzdata fontconfig freetype \
    font-inter wget bash libstdc++ libx11 libxrender libxext

# Layer 4: Install wkhtmltopdf dependencies (rarely changes)
RUN mv /etc/apk/repositories /etc/apk/repositories.bak && \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.14/main" >> /etc/apk/repositories && \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.14/community" >> /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache --upgrade gcc g++ gettext && \
    apk add --no-cache wkhtmltopdf xvfb && \
    mv /etc/apk/repositories.bak /etc/apk/repositories

# Layer 5: Install fonts (rarely changes)
RUN mkdir fonts \
    && wget https://storage.googleapis.com/pg-regression-test/Inter.zip -O /app/fonts/Inter.zip \
    && unzip -d /usr/share/fonts/Inter /app/fonts/Inter.zip \
    && fc-cache -f -v \
    && rm -f -R /app/fonts

# Switch to non-root user for security
USER user

# Copy artifacts from builder
COPY --from=builder /app/backend-portal /app/backend-portal
COPY --from=builder /app/backend-portal_debug /app/backend-portal_debug
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/templates /app/templates
COPY --from=builder /go/bin/dlv /app/dlv

# Expose application port
EXPOSE 3000

# Set CMD
CMD [ "/bin/sh", "-c", "/app/backend-portal $mode --config /app/env/.config.yaml --secret /app/env/.secret.yaml" ]

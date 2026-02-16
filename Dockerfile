# ---------- Build Stage ----------
    FROM golang:1.23-alpine AS builder

    RUN apk add --no-cache gcc musl-dev jq

    
    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    ENV CGO_ENABLED=1
    
    RUN go build -o timetravel .
    
    
    # ---------- Runtime Stage ----------
    FROM alpine:latest
    
    RUN apk add --no-cache ca-certificates sqlite-libs
    
    # 🔥 Set working directory to /app
    WORKDIR /app
    
    # 🔥 Create db folder INSIDE container
    RUN mkdir -p /app/db
    
    COPY --from=builder /app/timetravel .
    COPY --from=builder /app/conf ./conf
    COPY --from=builder /app/script ./script
    
    EXPOSE 8080
    
    CMD ["./timetravel"]
    
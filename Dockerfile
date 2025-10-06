# Étape 1 : Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxcontrol ./cmd/bot

# Étape 2 : Runtime
FROM alpine:latest

# Installer les certificats CA (nécessaire pour HTTPS)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copier le binaire depuis l'étape de build
COPY --from=builder /app/proxcontrol .

ENV TZ=Europe/Paris

CMD ["./proxcontrol"]


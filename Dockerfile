# Audit Proxy Gateway Dockerfile

FROM golang:1.22

# Çalışma dizinini /app olarak ayarla
WORKDIR /app

# go mod ve go.sum dosyalarını kopyala
COPY go.mod go.sum ./

# Bağımlılıkları indir
RUN go mod download

# Kaynak kodu konteynere kopyala
COPY . .

# Uygulamayı derle
RUN go build -o audit-proxy-gateway ./cmd/main.go

# 8081 portunu dışa aç
EXPOSE 8081

# Çalıştır komutunu belirt
CMD ["./audit-proxy-gateway"]

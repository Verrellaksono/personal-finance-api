# STAGE 1: Build binary aplikasi Go
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# Salin modul definisi dan unduh dependensi terlebih dahulu agar terkena layer cache
COPY go.mod go.sum ./
RUN go mod download

# Salin seluruh kode sumber proyek
COPY . .

# Kompilasi aplikasi menjadi binary statis (tanpa dependensi pustaka C dinamis)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api



# STAGE 2: Image produksi yang sangat ramping dan aman
FROM alpine:latest

WORKDIR /app

# Pasang sertifikasi SSL publik dan data zona waktu standar
RUN apk --no-cache add ca-certificates tzdata

# Salin binary hasil kompilasi dari stage builder
COPY --from=builder /app/api /app/api

# Buka port :8080 agar dapat diakses dari luar kontainer
EXPOSE 8080

# Jalankan binary saat kontainer dinyalakan
CMD ["/app/api"]
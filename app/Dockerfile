# --- 빌드 스테이지 ---
FROM golang:1.25-alpine AS build

WORKDIR /src

# 의존성 캐시 레이어 (go.mod/go.sum만 먼저 복사)
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사 후 정적 바이너리 빌드
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notiflex-api .

# --- 실행 스테이지 (scratch) ---
FROM scratch

COPY --from=build /notiflex-api /notiflex-api

EXPOSE 8080
ENTRYPOINT ["/notiflex-api"]

module github.com/modami/be-payment-service

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang-migrate/migrate/v4 v4.18.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.7.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/sony/gobreaker v1.0.0
	github.com/spf13/viper v1.19.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

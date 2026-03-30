package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Kafka        KafkaConfig
	GRPC         GRPCConfig
	VNPay        VNPayConfig
	MoMo         MoMoConfig
	ZaloPay      ZaloPayConfig
	BankTransfer BankTransferConfig
	JWT          JWTConfig
	App          AppConfig
}

type AppConfig struct {
	Env string
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	Name         string
	User         string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

// DSN returns the PostgreSQL data source name.
func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
}

type GRPCConfig struct {
	Port string
}

type VNPayConfig struct {
	TMNCode     string
	HashSecret  string
	PaymentURL  string
	ReturnURL   string
	IPNURL      string
	QueryURL    string
}

type MoMoConfig struct {
	PartnerCode string
	AccessKey   string
	SecretKey   string
	APIEndpoint string
	ReturnURL   string
	NotifyURL   string
}

type ZaloPayConfig struct {
	AppID       string
	Key1        string
	Key2        string
	CreateURL   string
	QueryURL    string
	RefundURL   string
	CallbackURL string
	ReturnURL   string
}

type BankTransferConfig struct {
	BankName      string
	AccountNumber string
	AccountName   string
	Branch        string
}

type JWTConfig struct {
	Secret string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read .env file — ignore error if not present (env vars may be set directly).
	_ = viper.ReadInConfig()

	cfg := &Config{
		App: AppConfig{
			Env: viper.GetString("APP_ENV"),
		},
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:         viper.GetString("DB_HOST"),
			Port:         viper.GetString("DB_PORT"),
			Name:         viper.GetString("DB_NAME"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			SSLMode:      viper.GetString("DB_SSLMODE"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
			GroupID: viper.GetString("KAFKA_GROUP_ID"),
		},
		GRPC: GRPCConfig{
			Port: viper.GetString("GRPC_PORT"),
		},
		VNPay: VNPayConfig{
			TMNCode:    viper.GetString("VNPAY_TMN_CODE"),
			HashSecret: viper.GetString("VNPAY_HASH_SECRET"),
			PaymentURL: viper.GetString("VNPAY_PAYMENT_URL"),
			ReturnURL:  viper.GetString("VNPAY_RETURN_URL"),
			IPNURL:     viper.GetString("VNPAY_IPN_URL"),
			QueryURL:   viper.GetString("VNPAY_QUERY_URL"),
		},
		MoMo: MoMoConfig{
			PartnerCode: viper.GetString("MOMO_PARTNER_CODE"),
			AccessKey:   viper.GetString("MOMO_ACCESS_KEY"),
			SecretKey:   viper.GetString("MOMO_SECRET_KEY"),
			APIEndpoint: viper.GetString("MOMO_API_ENDPOINT"),
			ReturnURL:   viper.GetString("MOMO_RETURN_URL"),
			NotifyURL:   viper.GetString("MOMO_NOTIFY_URL"),
		},
		ZaloPay: ZaloPayConfig{
			AppID:       viper.GetString("ZALOPAY_APP_ID"),
			Key1:        viper.GetString("ZALOPAY_KEY1"),
			Key2:        viper.GetString("ZALOPAY_KEY2"),
			CreateURL:   viper.GetString("ZALOPAY_CREATE_URL"),
			QueryURL:    viper.GetString("ZALOPAY_QUERY_URL"),
			RefundURL:   viper.GetString("ZALOPAY_REFUND_URL"),
			CallbackURL: viper.GetString("ZALOPAY_CALLBACK_URL"),
			ReturnURL:   viper.GetString("ZALOPAY_RETURN_URL"),
		},
		BankTransfer: BankTransferConfig{
			BankName:      viper.GetString("BANK_NAME"),
			AccountNumber: viper.GetString("BANK_ACCOUNT_NUMBER"),
			AccountName:   viper.GetString("BANK_ACCOUNT_NAME"),
			Branch:        viper.GetString("BANK_BRANCH"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
		},
	}

	// Set defaults.
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.GRPC.Port == "" {
		cfg.GRPC.Port = "9090"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.App.Env == "" {
		cfg.App.Env = "development"
	}

	return cfg, nil
}

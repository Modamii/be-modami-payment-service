package config

import (
	"fmt"
	"os"
	"path/filepath"
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

// Load reads configuration from a YAML file and environment variables (env overrides file).
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("app.env", "development")
	v.SetDefault("server.port", "8080")
	v.SetDefault("grpc.port", "9090")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 10)
	v.SetDefault("kafka.brokers", []string{})
	v.SetDefault("kafka.groupId", "payment-service")

	// File path (override via CONFIG_FILE)
	cfgFile := v.GetString("config.file")
	if cfgFile == "" {
		cfgFile = "config/config.yml"
	}
	if envPath := strings.TrimSpace(strings.ReplaceAll(getenv("CONFIG_FILE"), "\\", "/")); envPath != "" {
		cfgFile = envPath
	}

	v.SetConfigFile(cfgFile)
	switch strings.ToLower(filepath.Ext(cfgFile)) {
	case ".yml", ".yaml":
	default:
		return nil, fmt.Errorf("unsupported config file type: %s", cfgFile)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", cfgFile, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Backward-compatible: if kafka.brokers provided as comma string via env
	if brokers := strings.TrimSpace(getenv("KAFKA_BROKERS")); brokers != "" {
		cfg.Kafka.Brokers = splitCSV(brokers)
	}
	if group := strings.TrimSpace(getenv("KAFKA_GROUP_ID")); group != "" {
		cfg.Kafka.GroupID = group
	}

	return &cfg, nil
}

func getenv(k string) string { return strings.TrimSpace(os.Getenv(k)) }

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

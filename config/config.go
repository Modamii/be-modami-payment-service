package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Database      DatabaseConfig
	Redis         RedisConfig
	Kafka         KafkaConfig
	GRPC          GRPCConfig
	VNPay         VNPayConfig
	MoMo          MoMoConfig
	ZaloPay       ZaloPayConfig
	BankTransfer  BankTransferConfig
	JWT           JWTConfig
}

type AppConfig struct {
	Name            string `mapstructure:"name"`
	Version         string `mapstructure:"version"`
	Environment     string `mapstructure:"environment"`
	Debug           bool   `mapstructure:"debug"`
	Port            int    `mapstructure:"port"`
	Host            string `mapstructure:"host"`
	SwaggerHost     string `mapstructure:"swagger_host"`
	ShutdownTimeout  string   `mapstructure:"shutdown_timeout"`
	ReadTimeout      string   `mapstructure:"read_timeout"`
	WriteTimeout     string   `mapstructure:"write_timeout"`
	IdleTimeout      string   `mapstructure:"idle_timeout"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

func (a AppConfig) ListenAddr() string {
	host := strings.TrimSpace(a.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := a.Port
	if port <= 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func (a AppConfig) GetShutdownTimeout() time.Duration {
	d, err := time.ParseDuration(a.ShutdownTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetReadTimeout() time.Duration {
	d, err := time.ParseDuration(a.ReadTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetWriteTimeout() time.Duration {
	d, err := time.ParseDuration(a.WriteTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetIdleTimeout() time.Duration {
	d, err := time.ParseDuration(a.IdleTimeout)
	if err != nil {
		return 120 * time.Second
	}
	return d
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

type ObservabilityConfig struct {
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
	Environment    string `mapstructure:"environment"`
	LogLevel       string `mapstructure:"log_level"`
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	OTLPInsecure   bool   `mapstructure:"otlp_insecure"`
}

// Load reads configuration from a YAML file and environment variables (env overrides file).
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("app.name", "be-modami-payment-service")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", false)
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.swagger_host", "localhost:8080")
	v.SetDefault("app.shutdown_timeout", "30s")
	v.SetDefault("app.read_timeout", "30s")
	v.SetDefault("app.write_timeout", "30s")
	v.SetDefault("app.idle_timeout", "120s")
	v.SetDefault("app.allow_credentials", true)
	v.SetDefault("app.allowed_origins", []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080", "http://localhost:8081"})
	v.SetDefault("grpc.port", "9090")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 10)
	v.SetDefault("kafka.brokers", []string{})
	v.SetDefault("kafka.groupId", "payment-service")
	v.SetDefault("observability.service_name", "payment-service")
	v.SetDefault("observability.service_version", "1.0.0")
	v.SetDefault("observability.environment", "development")
	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.otlp_insecure", true)

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

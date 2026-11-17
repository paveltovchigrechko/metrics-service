package config

import (
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

// Default constants for command line flags.
const (
	serverAddressFlag   = "a"
	storeIntervalFlag   = "i"
	fileStoragePathFlag = "f"
	restoreFlag         = "r"
	databaseDSNFlag     = "d"
	secretKeyFlag       = "k"
	auditFilePathFlag   = "audit-file"
	auditURLFlag        = "audit-url"

	defaultAddress         = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.json"
	defaultDatabaseDSN     = ""
	defaultSecretKey       = ""
)

var (
	errEmptyFileStoragePath = errors.New("environment variable for file storage path is an empty string")
	errEmptyServerAddress   = errors.New("environment variable for server address is an empty string")
)

// ServerConfig holds the server parameters.
type ServerConfig struct {
	ServerAddress   string        // Address with port to serve the server.
	StoreInterval   time.Duration // Interval for storing the server metrics. If value is 0, server saves metrics after each update.
	FileStoragePath string        // Local file address to store server metrics.
	Restore         bool          // Indicates if the server restores the metrics saved on FileStoragePath on start.
	DatabaseDSN     string        // Address to connect to database.
	Key             string
	AuditFilePath   string // File path to store audit logs.
	AuditURL        string // URL to send audit logs.
	CryptoKey       string // Path to the PEM file with the private key to decrypt agent requests.
}

// envServerCongig stores the environment variables. All fields are pointers to distinguish if a value was set.
type envServerConfig struct {
	ServerAddress   *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
	AuditFilePath   *string `env:"AUDIT_FILE"`
	AuditURL        *string `env:"AUDIT_URL"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
	Config          *string `env:"CONFIG"`
}

// SetServerConfig returns the server configuration for args.
// Sources are applied from lowest priority to highest: the JSON config file,
// then flags that were actually passed, then environment variables.
// The file path comes from -c/-config or CONFIG, and CONFIG wins when both are set.
func SetServerConfig(args []string) (*ServerConfig, error) {
	envCfg, err := parseEnvServerConfig()
	if err != nil {
		return nil, err
	}

	flagCfg, explicit, flagPath, err := createServerFlagConfig(args)
	if err != nil {
		return nil, err
	}

	var fileCfg fileServerConfig
	if err := loadFileConfig(configFilePath(flagPath, envCfg.Config), &fileCfg); err != nil {
		return nil, err
	}
	if err := applyServerFileConfig(flagCfg, &fileCfg, explicit); err != nil {
		return nil, err
	}

	cfg := mergeServerConfigs(envCfg, flagCfg)

	// Validate interval, 0 is valid here.
	if cfg.StoreInterval < 0 {
		return nil, errIncorrectInterval
	}

	return cfg, nil
}

// mergeServerConfigs merges envServerConfig and ServerConfig with environment configuration precedence.
func mergeServerConfigs(envCfg *envServerConfig, flagCfg *ServerConfig) *ServerConfig {
	// Set server address
	if envCfg.ServerAddress != nil {
		flagCfg.ServerAddress = *envCfg.ServerAddress
	}

	// Set store interval
	if envCfg.StoreInterval != nil {
		storeInterval := time.Duration(*envCfg.StoreInterval) * time.Second
		flagCfg.StoreInterval = storeInterval
	}

	// Set file path to store metrics
	if envCfg.FileStoragePath != nil {
		flagCfg.FileStoragePath = *envCfg.FileStoragePath
	}

	// Set restore option
	if envCfg.Restore != nil {
		flagCfg.Restore = *envCfg.Restore
	}

	// Set database address
	if envCfg.DatabaseDSN != nil {
		flagCfg.DatabaseDSN = *envCfg.DatabaseDSN
	}

	// Set key
	if envCfg.Key != nil {
		flagCfg.Key = *envCfg.Key
	}

	// Set audit file path
	if envCfg.AuditFilePath != nil {
		flagCfg.AuditFilePath = *envCfg.AuditFilePath
	}

	// Set audit URL
	if envCfg.AuditURL != nil {
		flagCfg.AuditURL = *envCfg.AuditURL
	}

	// Set crypto key path
	if envCfg.CryptoKey != nil {
		flagCfg.CryptoKey = strings.TrimSpace(*envCfg.CryptoKey)
	}

	return flagCfg
}

// createServerFlagConfig parses args into a ServerConfig. explicit contains the
// names of flags present on the command line, and configPath is the -c/-config
// value. Parsed values are not validated.
func createServerFlagConfig(args []string) (cfg *ServerConfig, explicit map[string]bool, configPath string, err error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	address := fs.String(serverAddressFlag, defaultAddress, "Metrics server HTTP address and port")
	storeInterval := fs.Int(storeIntervalFlag, defaultStoreInterval, "Metrics store interval in seconds")
	fileStoragePath := fs.String(fileStoragePathFlag, defaultFileStoragePath, "File name to store metrics")
	restore := fs.Bool(restoreFlag, true, "Indicate if server should restore metrics from the storage file")
	databaseDSN := fs.String(databaseDSNFlag, defaultDatabaseDSN, "Database data source name")
	key := fs.String(secretKeyFlag, defaultSecretKey, "Secret key to encrypt data")
	auditFilePath := fs.String(auditFilePathFlag, "", "File path to store audit logs")
	auditURL := fs.String(auditURLFlag, "", "URL to send audit logs")
	cryptoKey := fs.String(cryptoKeyFlag, "", "Path to the PEM file with the private key to decrypt agent requests")
	configPathFlag := fs.String(configFlagShort, "", "Path to the JSON configuration file")
	fs.StringVar(configPathFlag, configFlagLong, "", "Path to the JSON configuration file")

	if err := fs.Parse(args); err != nil {
		return nil, nil, "", err
	}

	explicit = map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	storeInt := time.Duration(*storeInterval) * time.Second
	cleanedAuditFilePath := strings.TrimSpace(*auditFilePath)
	cleanedAuditURL := strings.TrimSpace(*auditURL)

	cfg = &ServerConfig{
		ServerAddress:   *address,
		StoreInterval:   storeInt,
		FileStoragePath: *fileStoragePath,
		Restore:         *restore,
		DatabaseDSN:     *databaseDSN,
		Key:             *key,
		AuditFilePath:   cleanedAuditFilePath,
		AuditURL:        cleanedAuditURL,
		CryptoKey:       strings.TrimSpace(*cryptoKey),
	}

	return cfg, explicit, strings.TrimSpace(*configPathFlag), nil
}

// parseEnvServerConfig parses the environment variables, validates string values for empty values, and returns envServerConfig.
func parseEnvServerConfig() (*envServerConfig, error) {
	envCfg := &envServerConfig{}

	err := env.Parse(envCfg)
	if err != nil {
		return nil, err
	}

	// Treat empty environment variables as error.
	// Don't check DATABASE_DSN and KEY: it's OK to have them empty.
	if envCfg.FileStoragePath != nil && strings.TrimSpace(*envCfg.FileStoragePath) == "" {
		return nil, errEmptyFileStoragePath
	}

	if envCfg.ServerAddress != nil && strings.TrimSpace(*envCfg.ServerAddress) == "" {
		return nil, errEmptyServerAddress
	}

	if envCfg.AuditFilePath != nil && strings.TrimSpace(*envCfg.AuditFilePath) == "" {
		*envCfg.AuditFilePath = ""
	}

	if envCfg.AuditURL != nil && strings.TrimSpace(*envCfg.AuditURL) == "" {
		*envCfg.AuditURL = ""
	}

	return envCfg, nil
}

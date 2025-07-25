package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/netip"
	"os"

	"dario.cat/mergo"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/ip"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/caarlos0/env/v11"
)

type CommonConfig struct {
	LogLvl        string `env:"LOG_LVL" json:"log_lvl"`
	Key           string `env:"KEY" json:"hmac_key"`
	CryptoKeyPath string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config        string `env:"CONFIG"`
	IsDev         bool   `env:"IS_DEV" json:"dev"`
}

type ServerConfig struct {
	PrivateKey      *rsa.PrivateKey
	TrustedNetwork  *netip.Prefix
	Addr            string `env:"ADDRESS" json:"address"`
	GRPCAddr        string `env:"GRPC_ADDR" json:"grpc_addr"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	CommonConfig
	StoreInterval int  `env:"STORE_INTERVAL" envDefault:"-1" json:"store_interval"`
	Restore       bool `env:"RESTORE" json:"restore"`
	SyncSave      bool
}

type AgentConfig struct {
	PublicKey *rsa.PublicKey
	IP        string
	Addr      string `env:"ADDRESS" json:"address"`
	CommonConfig
	ReportInterval int  `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int  `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit      int  `env:"RATE_LIMIT" json:"rate_limit"`
	IsHTTPS        bool `env:"IS_HTTPS" json:"https"`
	IsGRPC         bool `env:"IS_GRPC" json:"is_grpc"`
}

var defaultCommon = CommonConfig{
	LogLvl:        "debug",
	Key:           "",
	CryptoKeyPath: "",
	IsDev:         false,
	Config:        "",
}

var defaultServer = ServerConfig{
	CommonConfig:    defaultCommon,
	FileStoragePath: "store.json",
	Addr:            "localhost:8080",
	GRPCAddr:        ":3200",
	DatabaseDSN:     "",
	StoreInterval:   -1,
	Restore:         false,
	SyncSave:        false,
}

var defaultAgent = AgentConfig{
	CommonConfig:   defaultCommon,
	Addr:           "localhost:8080",
	ReportInterval: 10,
	PollInterval:   2,
	RateLimit:      5,
	IsHTTPS:        false,
	IsGRPC:         false,
}

func loadCommonFlags(flSet *flag.FlagSet, cfg *CommonConfig) {
	flSet.StringVar(&cfg.LogLvl, "log", "", "logging level")
	flSet.StringVar(&cfg.Key, "k", "", "key for SHA256")
	flSet.BoolVar(&cfg.IsDev, "dev", false, "is development")
	flSet.StringVar(&cfg.CryptoKeyPath, "crypto-key", "",
		"Path to the PEM-encoded RSA key (public for agent, private for server)")
	flSet.StringVar(&cfg.Config, "config", "", "path to JSON config file Usage: -config/-c")
	flSet.StringVar(&cfg.Config, "c", "", "path to JSON config file Usage: -config/-c")
}

func loadJSON[T *ServerConfig | *AgentConfig](path string, cfg T) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonBytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, cfg)
	if err != nil {
		return err
	}

	return nil
}

func mergeConfigs[T ServerConfig | AgentConfig](cfg *T, jsonCfg, flCfg, envVars T) error {
	if err := mergo.Merge(cfg, jsonCfg, mergo.WithOverride); err != nil {
		return err
	}
	if err := mergo.Merge(cfg, flCfg, mergo.WithOverride); err != nil {
		return err
	}
	if err := mergo.Merge(cfg, envVars, mergo.WithOverride); err != nil {
		return err
	}

	return nil
}

func loadServerFlags(flSet *flag.FlagSet, fl *ServerConfig) {
	flSet.StringVar(&fl.Addr, "a", "", "address for run server")
	flSet.IntVar(&fl.StoreInterval, "i", 0, "time is seconds to save db to store(file)")
	flSet.StringVar(&fl.FileStoragePath, "f", "", "path to save store")
	flSet.BoolVar(&fl.Restore, "r", false, "restore db from file")
	flSet.StringVar(&fl.DatabaseDSN, "d", "", "Postgres database DSN")
	flSet.StringVar(&fl.TrustedSubnet, "t", "", "Trusted subnet in CIDR notation (e.g., 192.168.1.0/24)")
	flSet.StringVar(&fl.GRPCAddr, "g", "", "GRPC address")
	loadCommonFlags(flSet, &fl.CommonConfig)
}

func parsePrivateKey(key string) *rsa.PrivateKey {
	if key == "" {
		return nil
	}

	pemBlock, err := crypto.ParsePEMFile(key)
	if err != nil {
		panic("error while parse pem\n" + err.Error())
	}

	priv, err := x509.ParsePKCS1PrivateKey(pemBlock)
	if err != nil {
		panic("error while parse pem\n" + err.Error())
	}

	return priv
}

func LoadServerCfg(args ...string) ServerConfig {
	flCfg := ServerConfig{}
	flSet := flag.NewFlagSet("server", flag.ContinueOnError)
	loadServerFlags(flSet, &flCfg)
	flSet.Parse(args)

	var envCfg ServerConfig
	err := env.Parse(&envCfg)
	if err != nil {
		logger.Log.Error("Error env vars")
	}

	configPath := envCfg.Config
	if configPath == "" {
		configPath = flCfg.Config
	}

	var jsonCfg ServerConfig
	if configPath != "" {
		err = loadJSON(configPath, &jsonCfg)
		if err != nil {
			fmt.Println("Error while getting json config\n", err.Error())
		}
	}

	cfg := defaultServer
	err = mergeConfigs(&cfg, jsonCfg, flCfg, envCfg)
	if err != nil {
		panic("merge err\n" + err.Error())
	}

	if cfg.DatabaseDSN != "" {
		cfg.Restore = false
	} else {
		cfg.SyncSave = cfg.StoreInterval == 0
	}

	cfg.PrivateKey = parsePrivateKey(cfg.CryptoKeyPath)

	if cfg.TrustedSubnet != "" {
		network, err := netip.ParsePrefix(cfg.TrustedSubnet)
		if err != nil {
			panic(err)
		}
		cfg.TrustedNetwork = &network
	}

	return cfg
}

func loadAgentFlags(flSet *flag.FlagSet, fl *AgentConfig) {
	flSet.IntVar(&fl.ReportInterval, "r", 0, "interval for sending runtime metrics")
	flSet.IntVar(&fl.PollInterval, "p", 0, "interval for getting runtime metrics")
	flSet.IntVar(&fl.RateLimit, "l", 0, "rate limit requests")
	flSet.BoolVar(&fl.IsHTTPS, "s", false, "https true/false, default false")
	flSet.BoolVar(&fl.IsGRPC, "g", false, "metrics will be sent via GRPC")
	flSet.StringVar(&fl.Addr, "a", "", "server address (http/grpc)")
	loadCommonFlags(flSet, &fl.CommonConfig)
}

func parsePublicKey(key string) *rsa.PublicKey {
	if key == "" {
		return nil
	}

	pemBlock, err := crypto.ParsePEMFile(key)
	if err != nil {
		panic("error while parse pem\n" + err.Error())
	}

	pub, err := x509.ParsePKCS1PublicKey(pemBlock)
	if err != nil {
		panic("error while parse pem\n" + err.Error())
	}

	return pub
}

func LoadAgentCfg(args ...string) AgentConfig {
	flCfg := AgentConfig{}
	flSet := flag.NewFlagSet("agent", flag.ContinueOnError)
	loadAgentFlags(flSet, &flCfg)
	flSet.Parse(args)

	var envVars AgentConfig
	err := env.Parse(&envVars)
	if err != nil {
		logger.Log.Error("Error parse flags")
	}

	configPath := envVars.Config
	if configPath == "" {
		configPath = flCfg.Config
	}

	var jsonCfg AgentConfig
	if configPath != "" {
		err = loadJSON(configPath, &jsonCfg)
		if err != nil {
			fmt.Println("Error while getting json config\n", err.Error())
		}
	}

	cfg := defaultAgent
	mergeConfigs(&cfg, jsonCfg, flCfg, envVars)

	cfg.PublicKey = parsePublicKey(cfg.CryptoKeyPath)

	cfg.IP, err = ip.GetOutboundIP()
	if err != nil {
		panic(err)
	}

	return cfg
}

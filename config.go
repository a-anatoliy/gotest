package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config holds the application settings. JSON keys match the original
// cfg/db.cfg format, so existing config files keep working.
type Config struct {
	DBUser     string `json:"DB_USERNAME"`
	DBPassword string `json:"DB_PASSWORD"`
	DBHost     string `json:"DB_HOST"`
	DBPort     string `json:"DB_PORT"`
	DBName     string `json:"DB_NAME"`
	ListenAddr string `json:"LISTEN_ADDR"`
}

func defaultConfig() Config {
	return Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		ListenAddr: ":8080",
	}
}

// LoadConfig reads the JSON file at path (if it exists) and then applies
// environment variable overrides (DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT,
// DB_NAME, LISTEN_ADDR). Environment variables win, which suits containers.
func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			if err := json.Unmarshal(data, &cfg); err != nil {
				return cfg, fmt.Errorf("parse config %s: %w", path, err)
			}
		case errors.Is(err, os.ErrNotExist):
			// No file is fine: settings may come from the environment.
		default:
			return cfg, fmt.Errorf("read config %s: %w", path, err)
		}
	}

	overrides := map[string]*string{
		"DB_USERNAME": &cfg.DBUser,
		"DB_PASSWORD": &cfg.DBPassword,
		"DB_HOST":     &cfg.DBHost,
		"DB_PORT":     &cfg.DBPort,
		"DB_NAME":     &cfg.DBName,
		"LISTEN_ADDR": &cfg.ListenAddr,
	}
	for name, field := range overrides {
		if v, ok := os.LookupEnv(name); ok {
			*field = v
		}
	}

	if cfg.DBUser == "" || cfg.DBName == "" {
		return cfg, errors.New("DB_USERNAME and DB_NAME must be set (config file or environment)")
	}
	return cfg, nil
}

// DSN builds the MySQL connection string with the driver's own formatter,
// so special characters in the password survive a round trip.
func (c Config) DSN() string {
	m := mysql.NewConfig()
	m.User = c.DBUser
	m.Passwd = c.DBPassword
	m.Net = "tcp"
	m.Addr = c.DBHost + ":" + c.DBPort
	m.DBName = c.DBName
	m.ParseTime = true
	m.Timeout = 5 * time.Second
	m.Params = map[string]string{"charset": "utf8mb4"}
	return m.FormatDSN()
}

package config

import (
	"fmt"
	"regexp"
	"time"

	"github.com/caarlos0/env/v11"
)

var GithubPattern = regexp.MustCompile(`^https://github\.com/([\w-]+/[\w-]+)$`)

type Config struct {
	TelegramAPIKey     string        `env:"TELEGRAM_API_KEY,required"`
	SurveyPeriod       time.Duration `env:"SURVEY_PERIOD" envDefault:"3600s"`
	FetchingStepPeriod time.Duration `env:"FETCHING_STEP_PERIOD" envDefault:"60s"`
	DBName             string        `env:"DB_NAME" envDefault:"db.sqlite3"`
}

func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}

package config

import (
	"os"
	"regexp"
	"time"
)

var (
	TelegramAPIKey     = ""
	SurveyPeriod       = time.Hour
	FetchingStepPeriod = time.Minute
	GithubPattern      = regexp.MustCompile(`^https://github\.com/([\w-]+/[\w-]+)$`)
)

const (
	DBName                string = "db.sqlite3"
	envTelegramAPIKey     string = "TELEGRAM_API_KEY"
	envSurveyPeriodKey    string = "SURVEY_PERIOD"
	envFetchingStepPeriod string = "FETCHING_STEP_PERIOD"
)

func LoadConfigFromEnv() {
	// Access environmental variables
	TelegramAPIKey = os.Getenv(envTelegramAPIKey)

	surveyPeriodStr := os.Getenv(envSurveyPeriodKey)
	if surveyPeriodStr != "" {
		surveyPeriodInt, err := time.ParseDuration(surveyPeriodStr)
		if err == nil {
			SurveyPeriod = surveyPeriodInt
		}
	}

	fetchingStepPeriodStr := os.Getenv(envFetchingStepPeriod)
	if fetchingStepPeriodStr != "" {
		fetchingStepPeriodInt, err := time.ParseDuration(fetchingStepPeriodStr)
		if err == nil {
			FetchingStepPeriod = fetchingStepPeriodInt
		}
	}
}

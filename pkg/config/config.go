package config

import (
	"os"
	"regexp"
	"time"
)

type key int

const DBContextKey key = iota

var (
	// DBName Base DB name.
	DBName = "db.sqlite3"
	// SurveyPeriod Main timing config to prevent GitHub API limits.
	SurveyPeriod       = time.Hour
	FetchingStepPeriod = time.Minute
	// GithubPattern RegExp pattern for checking user input.
	GithubPattern = regexp.MustCompile(`^https://github\.com/([\w-]+/[\w-]+)$`)
)

func LoadConfigFromEnv() {
	// Access environmental variables
	surveyPeriodStr := os.Getenv("SURVEY_PERIOD")
	if surveyPeriodStr != "" {
		surveyPeriodInt, err := time.ParseDuration(surveyPeriodStr)
		if err == nil {
			SurveyPeriod = surveyPeriodInt
		}
	}

	fetchingStepPeriodStr := os.Getenv("FETCHING_STEP_PERIOD")
	if fetchingStepPeriodStr != "" {
		fetchingStepPeriodInt, err := time.ParseDuration(fetchingStepPeriodStr)
		if err == nil {
			FetchingStepPeriod = fetchingStepPeriodInt
		}
	}
}

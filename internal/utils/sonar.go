package utils

import (
	"os"
)

func GetSonarServerAccessInfo() (url string, token string) {
	url = os.Getenv("SONAR_URL")
	token = os.Getenv("SONAR_TOKEN")

	if url == "" {
		url = "http://sonarqube-test-deploy-service.default:9000"
	}

	if token == "" {
		token = "b543b9dffd8cfbca5d870bebdf18f66dc35794c7"
	}

	return url, token
}

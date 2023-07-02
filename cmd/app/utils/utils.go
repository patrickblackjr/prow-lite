package utils

import (
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v53/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
)

func InitGitHubClient() {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, 123456, 123456789, "/config/prow-lite-qa.2023-04-01.private-key.pem")

	if err != nil {
		log.Fatal(err)
	}

	config.Config.GitHubClient = github.NewClient(&http.Client{Transport: itr})
}

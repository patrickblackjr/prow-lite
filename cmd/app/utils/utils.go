package utils

import (
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v45/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
)

func InitGitHubClient() {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, 269804, 32477892, "../config/github-app.pem")

	if err != nil {
		log.Fatal(err)
	}

	config.Config.GitHubClient = github.NewClient(&http.Client{Transport: itr})
}

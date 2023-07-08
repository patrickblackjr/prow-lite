package utils

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v50/github"
	"github.com/patrickblackjr/prow-lite/cmd/app/config"
	log "github.com/sirupsen/logrus"
)

// InitGitHubClient creates a new GitHub client
// provided an installation ID
func InitGitHubClient(installationID int64) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, 269804, installationID, "prow-lite-qa.2023-07-03.private-key.pem")

	if err != nil {
		log.Fatal(err)
	}

	config.Config.GitHubClient = github.NewClient(&http.Client{Transport: itr})
}

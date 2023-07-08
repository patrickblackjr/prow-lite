package main

import (
	"net/http"

	"github.com/patrickblackjr/prow-lite/cmd/app/config"
	"github.com/patrickblackjr/prow-lite/cmd/app/plugins"
	"github.com/patrickblackjr/prow-lite/cmd/app/utils"
	log "github.com/sirupsen/logrus"

	"github.com/cbrgm/githubevents/githubevents"
)

func main() {

	log.SetLevel(log.DebugLevel)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// load application configurations
	if err := config.LoadConfig("./config"); err != nil {
		log.Panicf("invalid application configuration: %s", err)
	}

	handle := githubevents.New(config.Config.GitHubWebhookSecret)

	handle.OnIssueCommentCreated(
		plugins.NewResponder("this is a test"),
	)

	http.HandleFunc("/github/webhook", func(w http.ResponseWriter, r *http.Request) {
		utils.InitGitHubClient(config.Config.GitHubInstallationID)
		err := handle.HandleEventRequest(r)
		if err != nil {
			log.Println(err.Error())
		}
	})

	err := http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		log.Panic(err)
	}
}

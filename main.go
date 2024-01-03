package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/patrickblackjr/prow-lite/config"
	"github.com/patrickblackjr/prow-lite/pkg/handlers"
	"github.com/patrickblackjr/prow-lite/pkg/log"
	"github.com/sirupsen/logrus"
)

func main() {

	log.InitLogging()

	// load application configurations
	c, err := config.ReadConfig("./config/server.yaml")
	if err != nil {
		logrus.Panicf("invalid application configuration: %s", err)
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		c.GitHub,
		githubapp.WithClientUserAgent("prow-lite/0.0.1"),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
	)
	if err != nil {
		logrus.Panic(err)
	}

	// Create instances of your plugins
	assigner := &handlers.Assigner{
		ClientCreator: cc,
	}

	// Pass the plugins to the IssueCommentHandler
	issueCommentHandler := &handlers.IssueCommentHandler{
		ClientCreator: cc,
		Plugins:       []handlers.Plugin{assigner},
	}

	webhookHandler := githubapp.NewDefaultEventDispatcher(c.GitHub, issueCommentHandler)

	http.Handle(githubapp.DefaultWebhookRoute, webhookHandler)

	addr := fmt.Sprintf("%s:%d", c.Server.Address, c.Server.Port)
	logrus.Infof("Starting server on %s", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logrus.Panic(err)
	}
}

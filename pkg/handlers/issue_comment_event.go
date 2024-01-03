package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/go-github/v56/github"
	"github.com/palantir/go-githubapp/githubapp"
	log "github.com/sirupsen/logrus"
)

type Plugin interface {
	Handle(ctx context.Context, event *github.IssueCommentEvent) error
}

type IssueCommentHandler struct {
	ClientCreator githubapp.ClientCreator
	Plugins       []Plugin
}

func (h *IssueCommentHandler) Handles() []string {
	return []string{"issue_comment"}
}

func (h *IssueCommentHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.IssueCommentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Failed to unmarshal issue comment event payload: %v", err)
		return err
	}

	if event.GetAction() != "created" {
		log.Infof("Skipping issue comment event with action: %s", event.GetAction())
		return nil
	}

	for _, plugin := range h.Plugins {
		if err := plugin.Handle(ctx, &event); err != nil {
			log.Errorf("Failed to handle issue comment event with plugin: %v", err)
			return err
		}
	}

	return nil
}

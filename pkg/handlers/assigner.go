package handlers

import (
	"context"
	"regexp"

	"github.com/google/go-github/v56/github"
	"github.com/palantir/go-githubapp/githubapp"
	log "github.com/sirupsen/logrus"
)

type Assigner struct {
	ClientCreator githubapp.ClientCreator
}

func (a *Assigner) Handle(ctx context.Context, event *github.IssueCommentEvent) error {
	body := event.GetComment().GetBody()
	assignRegex := regexp.MustCompile(`/assign @(\w+)`)
	unassignRegex := regexp.MustCompile(`/unassign @(\w+)`)

	// Check if the comment contains an assign or unassign command
	if assignRegex.MatchString(body) {
		matches := assignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			assignee := match[1]
			err := a.assign(ctx, event, assignee)
			if err != nil {
				log.Errorf("Failed to assign: %v", err)
				return err
			}
		}
	}

	if unassignRegex.MatchString(body) {
		matches := unassignRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			assignee := match[1]
			err := a.unassign(ctx, event, assignee)
			if err != nil {
				log.Errorf("Failed to unassign: %v", err)
				return err
			}
		}

	}

	return nil
}

func (a *Assigner) assign(ctx context.Context, event *github.IssueCommentEvent, assignee string) error {
	client, err := a.ClientCreator.NewInstallationClient(githubapp.GetInstallationIDFromEvent(event))
	if err != nil {
		log.Errorf("Failed to create GitHub client: %v", err)
		return err
	}
	client.Reactions.CreateIssueCommentReaction(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Comment.ID, "+1")
	_, _, err = client.Issues.AddAssignees(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number, []string{assignee})
	if err != nil {
		log.Errorf("Failed to add assignees: %v", err)
		return err
	}

	log.Infof("Assignee added successfully")
	return nil
}

func (a *Assigner) unassign(ctx context.Context, event *github.IssueCommentEvent, assignee string) error {
	client, err := a.ClientCreator.NewInstallationClient(githubapp.GetInstallationIDFromEvent(event))
	if err != nil {
		log.Errorf("Failed to create GitHub client: %v", err)
		return err
	}

	issue, _, err := client.Issues.Get(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number)
	if err != nil {
		log.Errorf("Failed to get issue: %v", err)
		return err
	}

	// Check if the assignee is currently assigned to the issue
	isAssigned := false

	for _, assignee := range issue.Assignees {
		client.Reactions.CreateIssueCommentReaction(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Comment.ID, "+1")
		if assignee.GetLogin() == *assignee.Login {
			isAssigned = true
			break
		}
		log.Infof("Assignee is not currently assigned to the issue #%v", issue.Number)
	}

	if isAssigned {
		_, _, err = client.Issues.RemoveAssignees(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number, []string{assignee})
		if err != nil {
			log.Errorf("Failed to remove assignees: %v", err)
			return err
		}
		removedAssignees := []string{assignee}
		_, _, err = client.Issues.RemoveAssignees(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), *event.Issue.Number, removedAssignees)
		if err != nil {
			log.Errorf("Failed to remove assignees: %v", err)
			return err
		}
		log.Infof("Assignee removed successfully")
	}
	return nil
}

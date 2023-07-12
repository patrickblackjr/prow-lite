package assign

import (
	"regexp"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
)

const pluginName = "assign"

var (
	assignRegexp = regexp.MustCompile(`(?mi)^/(un)?assign(( @?[-\w]+?)*)\s*$`)
	// CCRegexp     = regexp.MustCompile(`(?mi)^/(un)?cc(( +@?[-/\w]+?)*)\s*$`)
)

type githubClient interface {
	AssignIssue(owner, repo string, number int, logins []string) error
	UnassignIssue(owner, repo string, number int, logins []string) error

	CreateComment(owner, repo string, number int, comment string) error
}

type handler struct {
	add    func(org, repo string, number int, users []string) error
	remove func(org, repo string, number int, users []string) error
	event  *github.IssueCommentEvent
	regexp *regexp.Regexp
	gc     githubClient
}

func newAssignHandler(e github.IssueCommentEvent, gc githubClient) *handler {
	return &handler{
		add:    gc.AssignIssue,
		remove: gc.UnassignIssue,
		event:  &e,
		regexp: assignRegexp,
		gc:     gc,
	}
}

// func AssignUsers(c *github.Client, event *github.IssueCommentEvent) {
// 	comment := &github.IssueComment{
// 		Body: github.String("I will assign someone based on that"),
// 	}

// 	ctx := context.Background()
// 	owner := *event.Repo.Owner.Login
// 	repo := *event.Repo.Name
// 	issueNumber := *event.Issue.Number

// 	c.Issues.AddAssignees(ctx, owner, repo, issueNumber)

// 	logrus.Printf("%s commented here %s", *event.Comment.User.Login, *event.Comment.HTMLURL)
// 	c.Issues.CreateComment(context.Background(), owner, repo, issueNumber, comment)
// }

func parseLogins(text string) []string {
	var parts []string
	for _, p := range strings.Split(text, " ") {
		t := strings.Trim(p, "@ ")
		if t == "" {
			continue
		}
		parts = append(parts, t)
	}
	return parts
}

func handle(h *handler) error {
	e := h.event
	org := *e.Repo.Owner.Login
	repo := *e.Repo.Name
	matches := h.regexp.FindAllStringSubmatch(*e.Comment.Body, -1)
	if matches == nil {
		return nil
	}
	users := make(map[string]bool)
	for _, re := range matches {
		add := re[1] != "un"
		if re[2] == "" {
			users[*e.Comment.User.Login] = add
		} else {
			for _, login := range parseLogins(re[2]) {
				users[login] = add
			}
		}
	}
	var toAdd, toRemove []string
	for login, add := range users {
		if add {
			toAdd = append(toAdd, login)
		} else {
			toRemove = append(toRemove, login)
		}
	}

	if len(toRemove) > 0 {
		logrus.Printf("Removing %s from %s/%s#%d: %v", org, repo, e.Issue.Number, toRemove)
		if err := h.remove(org, repo, *e.Issue.Number, toRemove); err != nil {
			return err
		}
	}
	if len(toAdd) > 0 {
		logrus.Printf("Adding %s to %s/%s#%d: %v", org, repo, e.Issue.Number, toAdd)
	}
	return nil
}

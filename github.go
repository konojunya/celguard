package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v77/github"
	"golang.org/x/oauth2"
)

const prlintMarker = "<!-- prlint:konojunya/prlint -->"

func GitHubClient(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}

func findExistingComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (*github.IssueComment, error) {
	opt := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opt)
		if err != nil {
			return nil, err
		}

		for _, c := range comments {
			if c.Body != nil && strings.Contains(*c.Body, prlintMarker) {
				return c, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return nil, nil
}

func upsertFailedComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int, body string) error {
	content := prlintMarker + "\n" + body

	existing, err := findExistingComment(ctx, client, owner, repo, prNumber)
	if err != nil {
		return err
	}
	if existing != nil {
		_, _, err := client.Issues.EditComment(ctx, owner, repo, existing.GetID(), &github.IssueComment{Body: &content})
		return err
	}
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: &content})
	return err
}

func deleteFailedComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int) error {
	opt := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		cs, resp, err := client.Issues.ListComments(ctx, owner, repo, prNumber, opt)
		if err != nil {
			return fmt.Errorf("list comments: %w", err)
		}
		for _, c := range cs {
			if c.Body != nil && strings.Contains(*c.Body, prlintMarker) {
				if _, err := client.Issues.DeleteComment(ctx, owner, repo, c.GetID()); err != nil {
					return fmt.Errorf("delete comment: %w", err)
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return nil
}

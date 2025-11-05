package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	testCases := map[string]struct {
		pr         *Event
		cfg        *Config
		wantErrMsg string
		wantErr    bool
	}{
		"Pass - title": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "Long Title",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.size() > 5",
					Error: "Title must be longer than 10 characters",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - title": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "Short",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.size() > 10",
					Error: "Title must be longer than 10 characters",
				},
			},
			wantErrMsg: "PRLint failed.\n[title] Title must be longer than 10 characters",
			wantErr:    true,
		},
		"Pass - body": {
			pr: &Event{
				PullRequest: PullRequest{
					Body: "not empty",
				},
			},
			cfg: &Config{
				"body": {
					CEL:   "value.size() > 0",
					Error: "Body must be not empty",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - body": {
			pr: &Event{
				PullRequest: PullRequest{
					Body: "",
				},
			},
			cfg: &Config{
				"body": {
					CEL:   "value.size() > 0",
					Error: "Body must be not empty",
				},
			},
			wantErrMsg: "PRLint failed.\n[body] Body must be not empty",
			wantErr:    true,
		},
		"Pass - author": {
			pr: &Event{
				PullRequest: PullRequest{
					User: PullRequestUser{
						Login: "konojunya",
					},
				},
			},
			cfg: &Config{
				"author": {
					CEL:   "value.size() > 0",
					Error: "Author must be not empty",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - author": {
			pr: &Event{
				PullRequest: PullRequest{
					User: PullRequestUser{
						Login: "",
					},
				},
			},
			cfg: &Config{
				"author": {
					CEL:   "value.size() > 0",
					Error: "Author must be not empty",
				},
			},
			wantErrMsg: "PRLint failed.\n[author] Author must be not empty",
			wantErr:    true,
		},
		"Pass - base_ref": {
			pr: &Event{
				PullRequest: PullRequest{
					Base: PullRequestBase{
						Ref: "main",
					},
				},
			},
			cfg: &Config{
				"base_ref": {
					CEL:   "value == 'main'",
					Error: "Base ref must be 'main'",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - base_ref": {
			pr: &Event{
				PullRequest: PullRequest{
					Base: PullRequestBase{
						Ref: "feature",
					},
				},
			},
			cfg: &Config{
				"base_ref": {
					CEL:   "value == 'main'",
					Error: "Base ref must be 'main'",
				},
			},
			wantErrMsg: "PRLint failed.\n[base_ref] Base ref must be 'main'",
			wantErr:    true,
		},
		"Pass - head_ref": {
			pr: &Event{
				PullRequest: PullRequest{
					Head: PullRequestHead{
						Ref: "feature",
					},
				},
			},
			cfg: &Config{
				"head_ref": {
					CEL:   "value == 'feature'",
					Error: "Head ref must be 'feature'",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - head_ref": {
			pr: &Event{
				PullRequest: PullRequest{
					Head: PullRequestHead{
						Ref: "main",
					},
				},
			},
			cfg: &Config{
				"head_ref": {
					CEL:   "value == 'feature'",
					Error: "Head ref must be 'feature'",
				},
			},
			wantErrMsg: "PRLint failed.\n[head_ref] Head ref must be 'feature'",
			wantErr:    true,
		},
		"Pass - labels": {
			pr: &Event{
				PullRequest: PullRequest{
					Labels: []PullRequestLabel{
						{Name: "feature"},
					},
				},
			},
			cfg: &Config{
				"labels": {
					CEL:   "'feature' in value",
					Error: "Labels must contain 'feature'",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - labels": {
			pr: &Event{
				PullRequest: PullRequest{
					Labels: []PullRequestLabel{
						{Name: "main"},
					},
				},
			},
			cfg: &Config{
				"labels": {
					CEL:   "'feature' in value",
					Error: "Labels must contain 'feature'",
				},
			},
			wantErrMsg: "PRLint failed.\n[labels] Labels must contain 'feature'",
			wantErr:    true,
		},
		"Pass - title matches conventional commits (feat)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "feat: add new feature",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Pass - title matches conventional commits (fix)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "fix: fix bug",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		"Failed - title does not match conventional commits (invalid prefix)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "invalid: example",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "PRLint failed.\n[title] PR title must follow conventional commits format",
			wantErr:    true,
		},
		"Failed - title does not match conventional commits (no colon)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "feat add new feature",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "PRLint failed.\n[title] PR title must follow conventional commits format",
			wantErr:    true,
		},
		"Failed - title does not match conventional commits (no space after colon)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "feat:new feature",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "PRLint failed.\n[title] PR title must follow conventional commits format",
			wantErr:    true,
		},
		"Failed - title does not match conventional commits (empty after colon)": {
			pr: &Event{
				PullRequest: PullRequest{
					Title: "feat: ",
				},
			},
			cfg: &Config{
				"title": {
					CEL:   "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')",
					Error: "PR title must follow conventional commits format",
				},
			},
			wantErrMsg: "PRLint failed.\n[title] PR title must follow conventional commits format",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := run(tc.pr, tc.cfg)
			if tc.wantErr {
				assert.EqualError(t, err, tc.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

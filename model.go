package main

type Rule struct {
	CEL   string `yaml:"cel"`
	Error string `yaml:"error"`
}

type Config map[string]Rule

type PullRequestUser struct {
	Login string `json:"login"`
}

type PullRequestBase struct {
	Ref string `json:"ref"`
}

type PullRequestHead struct {
	Ref string `json:"ref"`
}

type PullRequestLabel struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Number int                `json:"number"`
	Title  string             `json:"title"`
	Body   string             `json:"body"`
	User   PullRequestUser    `json:"user"`
	Base   PullRequestBase    `json:"base"`
	Head   PullRequestHead    `json:"head"`
	Labels []PullRequestLabel `json:"labels"`
}

type Event struct {
	PullRequest PullRequest `json:"pull_request"`
}

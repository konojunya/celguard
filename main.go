package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
)

func run(event *Event, cfg *Config) error {
	env, err := cel.NewEnv(
		ext.Strings(),
		cel.Variable("value", cel.DynType),
		cel.Variable("pr", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return fmt.Errorf("failed to create CEL environment: %w", err)
	}

	labels := make([]string, 0, len(event.PullRequest.Labels))
	for _, l := range event.PullRequest.Labels {
		labels = append(labels, l.Name)
	}

	prMap := map[string]any{
		"title":    event.PullRequest.Title,
		"body":     event.PullRequest.Body,
		"author":   event.PullRequest.User.Login,
		"base_ref": event.PullRequest.Base.Ref,
		"head_ref": event.PullRequest.Head.Ref,
		"labels":   labels,
	}

	valueByKey := map[string]any{
		"title":    event.PullRequest.Title,
		"body":     event.PullRequest.Body,
		"author":   event.PullRequest.User.Login,
		"base_ref": event.PullRequest.Base.Ref,
		"head_ref": event.PullRequest.Head.Ref,
		"labels":   labels,
	}

	var failures []string

	eval := func(key string, rule *Rule) error {
		if strings.TrimSpace(rule.CEL) == "" {
			return fmt.Errorf("CEL is empty")
		}
		ast, issues := env.Compile(rule.CEL)
		if issues.Err() != nil {
			return fmt.Errorf("failed to compile CEL: %w", issues.Err())
		}
		prg, err := env.Program(ast)
		if err != nil {
			return fmt.Errorf("failed to create program: %w", err)
		}
		val := valueByKey[key]
		out, _, err := prg.Eval(map[string]any{
			"value": val,
			"pr":    prMap,
		})
		if err != nil {
			return fmt.Errorf("failed to evaluate CEL: %w", err)
		}

		truth, ok := out.Value().(bool)
		if !ok && out.Type() == types.BoolType {
			truth = out == types.True
			ok = true
		}
		if !ok {
			return fmt.Errorf("CEL returned non-boolean value: %v", out.Value())
		}
		if !truth {
			msg := rule.Error
			if strings.TrimSpace(msg) == "" {
				msg = fmt.Sprintf("Rule '%s' failed", key)
			}
			failures = append(failures, fmt.Sprintf("[%s] %s", key, msg))
		}
		return nil
	}

	for key, rule := range *cfg {
		if err := eval(key, &rule); err != nil {
			failures = append(failures, fmt.Sprintf("[%s] %s", key, err.Error()))
			return fmt.Errorf("failed to evaluate rule '%s': %w", key, err)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("PRLint failed.\n%s", strings.Join(failures, "\n"))
	}

	return nil
}

func fail(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("::error title=prlint::%s\n", msg)
	os.Exit(2)
}

func main() {
	cfg, err := ReadConfig()
	if err != nil {
		fail("failed to read config: %v", err)
	}

	fmt.Println("config:", cfg)

	event, err := LoadEventFromGitHub()
	if err != nil {
		fail("failed to load event from GitHub: %v", err)
	}

	fmt.Println("event:", event)

	var outBuf strings.Builder
	err = func() error {
		err = run(event, cfg)
		if err != nil {
			outBuf.WriteString(err.Error())
		}

		return err
	}()

	if os.Getenv("GITHUB_EVENT_NAME") != "pull_request" {
		if err != nil {
			fmt.Println(outBuf.String())
			os.Exit(1)
		}
		fmt.Println("PRLint passed")
		return
	}

	ownerRepo := os.Getenv("GITHUB_REPOSITORY")
	slash := strings.IndexByte(ownerRepo, '/')
	owner, repo := ownerRepo[:slash], ownerRepo[slash+1:]
	ctx := context.Background()
	gh, ghErr := GitHubClient(ctx)
	// if github client creation fails, only print the error message and exit with status 1
	if ghErr != nil {
		fmt.Printf("failed to create GitHub client: %v", ghErr)
		if err != nil {
			fmt.Println(outBuf.String())
			os.Exit(1)
		}
		fmt.Println("PRLint passed")
		return
	}

	if err != nil {
		header := "### ❌ PRLint failed"
		body := fmt.Sprintf("%s\n\n```\n%s\n```", header, outBuf.String())
		if cErr := upsertFailedComment(ctx, gh, owner, repo, event.PullRequest.Number, body); cErr != nil {
			fmt.Printf("warn: comment upsert: %v\n", cErr)
		}
		fmt.Println(outBuf.String())
		os.Exit(1)
	}

	if dErr := deleteFailedComment(ctx, gh, owner, repo, event.PullRequest.Number); dErr != nil {
		fmt.Printf("warn: comment delete: %v\n", dErr)
	}
	fmt.Println("PRLint passed")
}

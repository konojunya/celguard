package main

import (
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

	if err := run(event, cfg); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("PRLint passed")
}

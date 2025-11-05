# celguard 🛡️

[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/konojunya/celguard/ci.yaml?branch=main)](https://github.com/konojunya/celguard/actions)

A GitHub Action that validates Pull Requests using [Common Expression Language (CEL)](https://github.com/google/cel-spec).

It reads rules from `.github/celguard.yaml` and ensures your PR title, body, branch, labels, etc. follow your team's conventions.

![](./screenshots/screen.png)

## ✨ Features

- 🔍 **CEL-based rules** — powerful and flexible expression syntax (regex, list operations, logic)
- ⚙️ **Custom config per repo** — define rules in `.github/celguard.yaml`
- 🧪 **Go implementation** — single binary, so fast

## 🚀 Quick Start

### 1. Create Configuration File

Create `.github/celguard.yaml` in your repository root.

```yaml
title:
  cel: "value.matches('^(feat|fix|docs|style|refactor|test|chore): .+')"
  error: PR title must follow conventional commits format
head_ref:
  cel: "!value.matches('^(main|master|develop)$')"
  error: Direct commits to main/master/develop are not allowed
labels:
  cel: "value.exists(l, l in ['bug', 'enhancement', 'documentation'])"
  error: At least one of [bug, enhancement, documentation] label is required
```

### 2. Add Workflow

Create `.github/workflows/celguard.yaml`.

```yaml
name: PR Lint
on:
  pull_request:
    types: [opened, synchronize, reopened, edited]

permissions:
  contents: read
  pull-requests: write

jobs:
  celguard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run PR Lint
        uses: konojunya/celguard@v1.1.2
```

## 📖 Configuration

### Supported Configuration Keys

You can define validation rules for the following keys:

| Key | Type | Description | Example |
|-----|------|-------------|---------|
| `title` | `string` | PR title | `"feat: add new feature"` |
| `body` | `string` | PR body | `"This PR adds..."` |
| `author` | `string` | PR author's GitHub username | `"konojunya"` |
| `base_ref` | `string` | Base branch name | `"main"` |
| `head_ref` | `string` | Head branch name (source branch of PR) | `"feature/new-feature"` |
| `labels` | `[]string` | Array of labels attached to PR | `["bug", "enhancement"]` |

### Configuration File Format

For each key, define rules in the following format:

```yaml
<key>:
  cel: <CEL expression>
  error: <error message>
```

- **`cel`**: CEL expression. The validation passes when this expression returns `true`
- **`error`**: Error message displayed when validation fails (optional)

## 🛠️ Development

### Testing

```bash
go test ./...
```

## 🤝 Contributing

Contributions are welcome! You can contribute by following these steps:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Contribution Guidelines

- Bug reports and feature requests are welcome in [Issues](https://github.com/konojunya/celguard/issues)
- Code changes should be submitted via Pull Request
- Please follow the existing code style
- When adding new features, please also add tests

## 📄 License

This project is licensed under the license specified in the [LICENSE](LICENSE) file.

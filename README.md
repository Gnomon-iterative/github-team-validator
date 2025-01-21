# GitHub Team Validator

A GitHub Action to validate team membership and repository status for namespace files.

## Features

- Validates if PR authors are members of specified GitHub teams
- Verifies existence and visibility of source code repositories
- Supports LGTM approval workflow from team members
- Provides helpful feedback via PR comments

## Usage

1. Add the workflow to your repository's `.github/workflows` directory:

```yaml
name: Team Validator

on:
  pull_request:
    branches:
      - main
    paths:
      - 'namespaces/**'

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Validate Team Membership
        uses: Gnomon-iterative/github-team-validator@v1.0.0
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          pr-number: ${{ github.event.pull_request.number }}
          organization: your-org-name
```

2. Create or update your namespace file with required annotations:

```yaml
metadata:
  annotations:
    team: your-team-name
    source-code: https://github.com/org/repo
```

## Inputs

| Input | Description | Required |
|-------|-------------|----------|
| `github-token` | GitHub token for API access | Yes |
| `pr-number` | Pull request number | Yes |
| `organization` | GitHub organization name | Yes |

## Development

### Prerequisites
- Go 1.20 or later
- Docker

### Building
```bash
go mod download
go build
```

### Running Tests
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of all notable changes.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

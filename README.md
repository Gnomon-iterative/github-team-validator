# GitHub Team Validator

A GitHub Action that validates team membership and repository status for namespace files in pull requests.

## Features

- Validates that PR author belongs to the specified team
- Verifies that source code repository exists and is public
- Checks namespace file annotations for team and source-code

## Usage

```yaml
name: Team Validator

on:
  pull_request:
    paths:
      - 'namespaces/**'

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

## Namespace File Format

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: example-namespace
  annotations:
    team: team-name
    source-code: https://github.com/org/repo
```

## Requirements

- GitHub token with read access to:
  - Organization teams
  - Repository contents
  - Pull requests
- Namespace files must include:
  - `team` annotation with valid team name
  - `source-code` annotation with public GitHub repository URL

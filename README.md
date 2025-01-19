# GitHub Team Validator

This GitHub Actions workflow validates namespace files in pull requests by checking:

1. Team membership validation
2. Repository existence and visibility checks

## How it works

The workflow runs automatically on pull requests when YAML files are modified. It performs the following checks:

### Team Membership Validation
- Extracts the team name from the `team` annotation in the namespace YAML
- Verifies that the PR author is a member of the specified team within your organization
- Fails if the user is not a member of the specified team

### Repository Validation
- Checks if the repository specified in the `source-code` annotation exists
- Verifies if the repository is public or private
- Comments on the PR if the repository is private
- Fails if the repository doesn't exist

## Example namespace.yaml

```yaml
metadata:
  annotations:
    team: engineering-team
    source-code: my-application
```

## Requirements

The workflow requires:

1. `GITHUB_TOKEN` with appropriate permissions to:
   - Read team memberships
   - Read repository information
   - Comment on pull requests

## Error Messages

The workflow will comment on the PR with specific error messages:

- When team membership validation fails
- When repository doesn't exist
- When repository is private
- When YAML is invalid
- When required annotations are missing

A success message is posted when all validations pass.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-21

### Added
- Initial release of GitHub Team Validator
- Team membership validation for PR authors
- Source code repository existence and visibility checks
- LGTM approval workflow from team members
- Docker-based GitHub Action implementation
- Multi-stage Docker build for minimal image size
- Comprehensive error handling and logging
- GitHub API integration using go-github v57
- YAML parsing with yaml.v3

### Dependencies
- Go 1.20
- github.com/google/go-github/v57
- golang.org/x/oauth2 v0.15.0
- gopkg.in/yaml.v3 v3.0.1

### Required Permissions
- `contents: read`
- `pull-requests: write`
- `issues: write`

[1.0.0]: https://github.com/Gnomon-iterative/github-team-validator/releases/tag/v1.0.0

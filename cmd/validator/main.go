package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3" // Using yaml.v3 instead of v2 for better performance and security
)

// Namespace represents the structure of the namespace YAML file
type Namespace struct {
	Metadata struct {
		Annotations struct {
			Team       string `yaml:"team"`
			SourceCode string `yaml:"source-code"`
		} `yaml:"annotations"`
	} `yaml:"metadata"`
}

// PRComment represents a GitHub PR comment
type PRComment struct {
	Body string `json:"body"`
}

// ValidationError is a custom error type for validation failures
type ValidationError struct {
	Message string
	File    string
}

func (e *ValidationError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("❌ Error in %s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("❌ Error: %s", e.Message)
}

// validateTeamMembership checks if the user is a member of the specified team
func validateTeamMembership(teamName, username, orgName, token string) error {
	// GitHub API endpoint for team membership
	url := fmt.Sprintf("https://api.github.com/orgs/%s/teams/%s/memberships/%s", orgName, teamName, username)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to check team membership: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ValidationError{
			Message: fmt.Sprintf("@%s is not a member of the team '%s'", username, teamName),
		}
	}

	return nil
}

// checkRepositoryStatus verifies if the repository exists and checks its visibility
func checkRepositoryStatus(repoName, orgName, token string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", orgName, repoName)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to check repository status: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &ValidationError{
			Message: fmt.Sprintf("Repository '%s' does not exist in the organization", repoName),
		}
	}

	var repoData struct {
		Private bool `json:"private"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoData); err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to parse repository data: %v", err)}
	}

	if repoData.Private {
		return &ValidationError{
			Message: fmt.Sprintf("Repository '%s' is private. Please ensure it's public for proper access", repoName),
		}
	}

	return nil
}

// commentOnPR posts a comment to the GitHub pull request
func commentOnPR(message, commentsURL, token string) error {
	comment := PRComment{Body: message}
	commentJSON, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %v", err)
	}

	req, err := http.NewRequest("POST", commentsURL, strings.NewReader(string(commentJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create comment, status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Get required environment variables
	token := os.Getenv("GITHUB_TOKEN")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	prAuthor := os.Getenv("GITHUB_ACTOR")
	orgName := os.Getenv("GITHUB_REPOSITORY_OWNER")

	// Read GitHub event data
	eventData, err := os.ReadFile(eventPath)
	if err != nil {
		log.Fatalf("Failed to read event file: %v", err)
	}

	var event struct {
		PullRequest struct {
			CommentsURL string `json:"comments_url"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Fatalf("Failed to parse event data: %v", err)
	}

	// Process changed YAML files from command line arguments
	hasErrors := false
	for _, filePath := range os.Args[1:] {
		if !strings.HasSuffix(filePath, ".yaml") && !strings.HasSuffix(filePath, ".yml") {
			continue
		}

		// Read and parse YAML file
		data, err := os.ReadFile(filePath)
		if err != nil {
			commentOnPR(fmt.Sprintf("Failed to read file %s: %v", filePath, err), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		var ns Namespace
		if err := yaml.Unmarshal(data, &ns); err != nil {
			commentOnPR(fmt.Sprintf("Invalid YAML in %s: %v", filePath, err), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Validate team annotation
		if ns.Metadata.Annotations.Team == "" {
			commentOnPR((&ValidationError{
				Message: "Team annotation is missing",
				File:    filePath,
			}).Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Validate source-code annotation
		if ns.Metadata.Annotations.SourceCode == "" {
			commentOnPR((&ValidationError{
				Message: "Source code repository annotation is missing",
				File:    filePath,
			}).Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Check team membership
		if err := validateTeamMembership(ns.Metadata.Annotations.Team, prAuthor, orgName, token); err != nil {
			commentOnPR(err.Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Check repository status
		if err := checkRepositoryStatus(ns.Metadata.Annotations.SourceCode, orgName, token); err != nil {
			commentOnPR(err.Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}
	}

	if !hasErrors {
		commentOnPR("✅ All team membership and repository validations passed!", event.PullRequest.CommentsURL, token)
	}

	if hasErrors {
		os.Exit(1)
	}
}

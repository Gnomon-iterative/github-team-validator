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
		Name string `yaml:"name"`
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
	log.Printf("Checking if user %s is a member of team %s in org %s", username, teamName, orgName)
	
	// GitHub API endpoint for team membership
	url := fmt.Sprintf("https://api.github.com/orgs/%s/teams/%s/memberships/%s", orgName, teamName, username)
	log.Printf("Making request to: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to check team membership: %v", err)}
	}
	defer resp.Body.Close()

	log.Printf("Got response status: %d", resp.StatusCode)
	
	if resp.StatusCode == 404 {
		return &ValidationError{Message: fmt.Sprintf("Team %s not found or user %s is not a member", teamName, username)}
	}
	if resp.StatusCode != 200 {
		return &ValidationError{Message: fmt.Sprintf("Failed to check team membership. Status: %d", resp.StatusCode)}
	}

	return nil
}

// checkRepositoryStatus verifies if the repository exists and checks its visibility
func checkRepositoryStatus(repoName, orgName, token string) error {
	log.Printf("Checking repository %s/%s", orgName, repoName)
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", orgName, repoName)
	log.Printf("Making request to: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &ValidationError{Message: fmt.Sprintf("Failed to check repository status: %v", err)}
	}
	defer resp.Body.Close()

	log.Printf("Got response status: %d", resp.StatusCode)
	
	if resp.StatusCode == 404 {
		return &ValidationError{Message: fmt.Sprintf("Repository %s/%s not found", orgName, repoName)}
	}
	if resp.StatusCode != 200 {
		return &ValidationError{Message: fmt.Sprintf("Failed to check repository status. Status: %d", resp.StatusCode)}
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
	// Enable debug logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Get required environment variables
	token := os.Getenv("GITHUB_TOKEN")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	prAuthor := os.Getenv("GITHUB_ACTOR")
	orgName := os.Getenv("GITHUB_REPOSITORY_OWNER")

	log.Printf("Starting validation with:")
	log.Printf("- PR Author: %s", prAuthor)
	log.Printf("- Organization: %s", orgName)
	log.Printf("- Event Path: %s", eventPath)
	log.Printf("- Token present: %v", token != "")

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
		log.Printf("Processing file: %s", filePath)
		
		if !strings.HasSuffix(filePath, ".yaml") && !strings.HasSuffix(filePath, ".yml") {
			log.Printf("Skipping non-YAML file: %s", filePath)
			continue
		}

		// Read and parse YAML file
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filePath, err)
			commentOnPR(fmt.Sprintf("Failed to read file %s: %v", filePath, err), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		var ns Namespace
		if err := yaml.Unmarshal(data, &ns); err != nil {
			log.Printf("Invalid YAML in %s: %v", filePath, err)
			commentOnPR(fmt.Sprintf("Invalid YAML in %s: %v", filePath, err), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		log.Printf("Parsed namespace:")
		log.Printf("- Team: %s", ns.Metadata.Annotations.Team)
		log.Printf("- Source Code: %s", ns.Metadata.Annotations.SourceCode)
		log.Printf("- Name: %s", ns.Metadata.Name)

		// Validate team annotation
		if ns.Metadata.Annotations.Team == "" {
			log.Printf("Team annotation is missing in %s", filePath)
			commentOnPR((&ValidationError{
				Message: "Team annotation is missing",
				File:    filePath,
			}).Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Validate source-code annotation
		if ns.Metadata.Annotations.SourceCode == "" {
			log.Printf("Source code repository annotation is missing in %s", filePath)
			commentOnPR((&ValidationError{
				Message: "Source code repository annotation is missing",
				File:    filePath,
			}).Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Check team membership
		if err := validateTeamMembership(ns.Metadata.Annotations.Team, prAuthor, orgName, token); err != nil {
			log.Printf("Team membership validation failed: %v", err)
			commentOnPR(err.Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}

		// Check repository status
		if err := checkRepositoryStatus(ns.Metadata.Annotations.SourceCode, orgName, token); err != nil {
			log.Printf("Repository validation failed: %v", err)
			commentOnPR(err.Error(), event.PullRequest.CommentsURL, token)
			hasErrors = true
			continue
		}
	}

	if !hasErrors {
		log.Printf("All validations passed!")
		commentOnPR("✅ All team membership and repository validations passed!", event.PullRequest.CommentsURL, token)
	}

	if hasErrors {
		os.Exit(1)
	}
}

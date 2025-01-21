package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

func main() {
	// Get inputs from environment variables
	token := os.Getenv("INPUT_GITHUB-TOKEN")
	prNumber := os.Getenv("INPUT_PR-NUMBER")
	orgName := os.Getenv("INPUT_ORGANIZATION")

	if token == "" || prNumber == "" || orgName == "" {
		fmt.Println("Error: Required inputs are missing")
		os.Exit(1)
	}

	// Create GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get PR details
	pr, _, err := client.PullRequests.Get(ctx, orgName, "cloud-cost", prNumber)
	if err != nil {
		fmt.Printf("Error getting PR details: %v\n", err)
		os.Exit(1)
	}

	// Get PR author
	author := pr.GetUser().GetLogin()

	// Get changed files
	files, _, err := client.PullRequests.ListFiles(ctx, orgName, "cloud-cost", pr.GetNumber(), &github.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting PR files: %v\n", err)
		os.Exit(1)
	}

	// Check each changed file
	for _, file := range files {
		if strings.HasPrefix(file.GetFilename(), "namespaces/") {
			// Get file content
			content, _, _, err := client.Repositories.GetContents(ctx, orgName, "cloud-cost", file.GetFilename(), &github.RepositoryContentGetOptions{
				Ref: pr.Head.GetSHA(),
			})
			if err != nil {
				fmt.Printf("Error getting file content: %v\n", err)
				continue
			}

			// Decode content
			fileContent, err := content.GetContent()
			if err != nil {
				fmt.Printf("Error decoding content: %v\n", err)
				continue
			}

			// Extract team name from annotations
			teamName := extractTeamName(fileContent)
			if teamName == "" {
				fmt.Printf("Error: No team annotation found in %s\n", file.GetFilename())
				os.Exit(1)
			}

			// Check if user is member of the team
			membership, _, err := client.Teams.GetTeamMembershipBySlug(ctx, orgName, teamName, author)
			if err != nil {
				fmt.Printf("Error checking team membership for %s in team %s: %v\n", author, teamName, err)
				os.Exit(1)
			}

			if membership.GetState() != "active" {
				fmt.Printf("Error: User %s is not an active member of team %s\n", author, teamName)
				os.Exit(1)
			}

			// Check source-code repository
			sourceRepo := extractSourceRepo(fileContent)
			if sourceRepo == "" {
				fmt.Printf("Error: No source-code annotation found in %s\n", file.GetFilename())
				os.Exit(1)
			}

			// Extract owner and repo from source-code URL
			parts := strings.Split(strings.TrimPrefix(sourceRepo, "https://github.com/"), "/")
			if len(parts) != 2 {
				fmt.Printf("Error: Invalid source-code URL format in %s\n", file.GetFilename())
				os.Exit(1)
			}

			// Check if repository exists and is public
			repo, _, err := client.Repositories.Get(ctx, parts[0], parts[1])
			if err != nil {
				fmt.Printf("Error: Source code repository does not exist or is not accessible: %v\n", err)
				os.Exit(1)
			}

			if repo.GetPrivate() {
				fmt.Printf("Error: Source code repository must be public\n")
				os.Exit(1)
			}
		}
	}

	fmt.Println("Validation successful!")
}

func extractTeamName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "team:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "team:"))
		}
	}
	return ""
}

func extractSourceRepo(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "source-code:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "source-code:"))
		}
	}
	return ""
}

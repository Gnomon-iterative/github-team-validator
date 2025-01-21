package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v53/github"
	"gopkg.in/yaml.v2"
	"golang.org/x/oauth2"
)

type NamespaceMetadata struct {
	Annotations map[string]string `yaml:"annotations"`
}

type Namespace struct {
	Metadata NamespaceMetadata `yaml:"metadata"`
}

func main() {
	token := os.Getenv("INPUT_GITHUB-TOKEN")
	prNumber := os.Getenv("INPUT_PR-NUMBER")
	org := os.Getenv("INPUT_ORGANIZATION")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Read the namespace.yaml file
	data, err := os.ReadFile("namespaces/test-namespace.yaml")
	if err != nil {
		fmt.Printf("Error reading namespace file: %v\n", err)
		os.Exit(1)
	}

	var ns Namespace
	if err := yaml.Unmarshal(data, &ns); err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	team := ns.Metadata.Annotations["team"]
	sourceCode := ns.Metadata.Annotations["source-code"]

	// Get PR author
	pr, _, err := client.PullRequests.Get(ctx, org, "cloud-cost", prNumber)
	if err != nil {
		fmt.Printf("Error getting PR: %v\n", err)
		os.Exit(1)
	}
	author := pr.User.GetLogin()

	// Check team membership
	isMember, _, err := client.Teams.GetTeamMembershipBySlug(ctx, org, team, author)
	if err != nil || isMember == nil {
		comment := fmt.Sprintf("@%s is not a member of team %s. Need LGTM from a team member to proceed.", author, team)
		_, _, err = client.Issues.CreateComment(ctx, org, "cloud-cost", prNumber, &github.IssueComment{Body: &comment})
		if err != nil {
			fmt.Printf("Error creating comment: %v\n", err)
		}
		os.Exit(1)
	}

	// Parse source code URL
	parts := strings.Split(sourceCode, "/")
	if len(parts) < 3 {
		comment := "Invalid source-code URL format"
		_, _, err = client.Issues.CreateComment(ctx, org, "cloud-cost", prNumber, &github.IssueComment{Body: &comment})
		os.Exit(1)
	}

	repoOwner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]

	// Check if repo exists and is public
	repo, _, err := client.Repositories.Get(ctx, repoOwner, repoName)
	if err != nil {
		comment := fmt.Sprintf("Repository %s does not exist", sourceCode)
		_, _, err = client.Issues.CreateComment(ctx, org, "cloud-cost", prNumber, &github.IssueComment{Body: &comment})
		os.Exit(1)
	}

	if repo.GetPrivate() {
		comment := fmt.Sprintf("Repository %s is private. Only public repositories are allowed", sourceCode)
		_, _, err = client.Issues.CreateComment(ctx, org, "cloud-cost", prNumber, &github.IssueComment{Body: &comment})
		os.Exit(1)
	}

	// Check for LGTM comment if user is not a team member
	if isMember == nil {
		opts := &github.ListOptions{PerPage: 100}
		comments, _, err := client.Issues.ListComments(ctx, org, "cloud-cost", prNumber, opts)
		if err != nil {
			fmt.Printf("Error getting comments: %v\n", err)
			os.Exit(1)
		}

		hasApproval := false
		for _, comment := range comments {
			if strings.Contains(comment.GetBody(), "LGTM") {
				commentAuthor := comment.User.GetLogin()
				isApprover, _, err := client.Teams.GetTeamMembershipBySlug(ctx, org, team, commentAuthor)
				if err == nil && isApprover != nil {
					hasApproval = true
					break
				}
			}
		}

		if !hasApproval {
			comment := fmt.Sprintf("Waiting for LGTM from a member of team %s", team)
			_, _, err = client.Issues.CreateComment(ctx, org, "cloud-cost", prNumber, &github.IssueComment{Body: &comment})
			os.Exit(1)
		}
	}

	fmt.Println("Validation successful!")
}

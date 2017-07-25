package webhook

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"github.com/deis/acid/pkg/acid"
	"github.com/google/go-github/github"
)

// State names for GitHub status
var (
	StatePending = "pending"
	StateFailure = "failure"
	StateError   = "error"
	StateSuccess = "success"
)

const StatusContext = "acid"

// ghClient gets a new GitHub client object.
//
// It authenticates with an OAUTH2 token.
func ghClient(token string) *github.Client {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	c := context.Background()
	tc := oauth2.NewClient(c, t)
	return github.NewClient(tc)
}

// setRepoStatus sets the status on a particular commit in a repo.
func setRepoStatus(repo, commit string, proj *acid.Project, status *github.RepoStatus) error {
	if proj.GitHubToken == "" {
		return fmt.Errorf("status update skipped because no GitHubToken exists on %s", proj.Name)
	}
	parts := strings.SplitN(proj.Repo.Name, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("project name %q is malformed", proj.Repo.Name)
	}
	c := context.Background()
	client := ghClient(proj.GitHubToken)
	_, _, err := client.Repositories.CreateStatus(
		c,
		parts[0],
		parts[1],
		commit,
		status)
	return err
}

// GetRepoStatus gets the Acid repository status.
// The ref can be a SHA or a branch or tag.
func GetRepoStatus(proj *acid.Project, ref string) (*github.RepoStatus, error) {
	c := context.Background()
	client := ghClient(proj.GitHubToken)
	parts := strings.SplitN(proj.Repo.Name, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("project name %q is malformed", proj.Repo.Name)
	}
	statii, _, err := client.Repositories.ListStatuses(c, parts[0], parts[1], ref, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, status := range statii {
		if *status.Context == StatusContext {
			return status, nil
		}
	}
	return nil, fmt.Errorf("no acid status found")
}

// GetLastCommit gets the last commit on the give reference (branch name or tag).
func GetLastCommit(proj *acid.Project, ref string) (string, error) {
	c := context.Background()
	client := ghClient(proj.GitHubToken)
	parts := strings.SplitN(proj.Repo.Name, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("project name %q is malformed", proj.Repo.Name)
	}
	sha, _, err := client.Repositories.GetCommitSHA1(c, parts[0], parts[1], ref, "")
	return sha, err
}

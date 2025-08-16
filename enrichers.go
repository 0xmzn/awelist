package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v74/github"
	"gitlab.com/gitlab-org/api/client-go"
)

type Enricher interface {
	Enrich() error
}

type RemoteRepo interface {
	Stars() int
	LastUpdate() time.Time
	Enrich() error
}

type githubClient struct {
	client  *github.Client
	once    sync.Once
	initErr error
}

var ghClientSingleton = &githubClient{}

func getGitHubClient() (*github.Client, error) {
	ghClientSingleton.once.Do(func() {
		token := os.Getenv("GITHUB_API_KEY")
		if token == "" {
			ghClientSingleton.initErr = CliErrorf(nil, "GITHUB_API_KEY environment variable is not set. A token is required for GitHub API calls.")
			return
		}
		ghClientSingleton.client = github.NewClient(http.DefaultClient).WithAuthToken(token)
	})

	if ghClientSingleton.initErr != nil {
		return nil, ghClientSingleton.initErr
	}

	if ghClientSingleton.client == nil {
		return nil, CliErrorf(nil, "GitHub client failed to initialize for an unknown reason.")
	}
	return ghClientSingleton.client, nil
}

type gitlabClient struct {
	client  *gitlab.Client
	once    sync.Once
	initErr error
}

var glClientSingleton = &gitlabClient{}

func getGitlabClient() (*gitlab.Client, error) {
	glClientSingleton.once.Do(func() {
		token := os.Getenv("GITLAB_API_KEY")
		if token == "" {
			glClientSingleton.initErr = CliErrorf(nil, "GITLAB_API_KEY environment variable is not set. A token is required for gitlab API calls.")
			return
		}
		// TODO: Deal with this error. It's currenty overriding the previous one
		client, err := gitlab.NewClient(token)

		glClientSingleton.client = client
		glClientSingleton.initErr = err
	})

	if glClientSingleton.initErr != nil {
		return nil, glClientSingleton.initErr
	}

	if glClientSingleton.client == nil {
		return nil, CliErrorf(nil, "gitlab client failed to initialize for an unknown reason.")
	}
	return glClientSingleton.client, nil
}

type githubRepo struct {
	url        string
	stars      int
	lastUpdate time.Time
}

type gitlabRepo struct {
	url        string
	stars      int
	lastUpdate time.Time
}

func NewRemoteRepo(urlStr string) RemoteRepo {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	if parsedURL.Host == "" || len(pathParts) < 2 || pathParts[0] == "" || pathParts[1] == "" {
		return nil
	}

	switch parsedURL.Host {
	case "github.com":
		if len(pathParts) != 2 {
			return nil
		}
		return &githubRepo{url: urlStr}
	case "gitlab.com":
		if len(pathParts) < 2 {
			return nil
		}
		return &gitlabRepo{url: urlStr}
	}

	return nil
}

func (repo *githubRepo) Stars() int {
	return repo.stars
}

func (repo *githubRepo) LastUpdate() time.Time {
	return repo.lastUpdate
}

func (repo *githubRepo) Enrich() error {
	parsedURL, err := url.Parse(repo.url)
	if err != nil {
		return CliErrorf(err, "invalid URL %q", repo.url)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return CliErrorf(err, "invalid GitHub URL format (expected owner/repo) found %q", repo.url)
	}

	owner := pathParts[0]
	repoName := pathParts[1]

	client, err := getGitHubClient()
	if err != nil {
		return CliErrorf(err, "failed to get GitHub client for %q", repo.url)
	}

	ctx := context.Background()
	ghRepo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return CliErrorf(err, "failed to fetch github repo details %q", repo.url)
	}

	if ghRepo.StargazersCount != nil {
		repo.stars = *ghRepo.StargazersCount
	}

	repo.stars = ghRepo.GetStargazersCount()
	defaultBranch := ghRepo.GetDefaultBranch()

	commit, _, err := client.Repositories.GetCommit(ctx, owner, repoName, defaultBranch, nil)

	if err != nil {
		return CliErrorf(err, "failed to fetch latest commit for default branch %q on %q", defaultBranch, repo.url)
	}

	if commit != nil && commit.Commit != nil && commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil {
		repo.lastUpdate = commit.Commit.Committer.Date.Time
	} else {
		if ghRepo.UpdatedAt != nil {
			repo.lastUpdate = ghRepo.UpdatedAt.Time
		}
	}

	return nil
}

func (repo *gitlabRepo) Stars() int {
	return repo.stars
}

func (repo *gitlabRepo) LastUpdate() time.Time {
	return repo.lastUpdate
}

func (repo *gitlabRepo) Enrich() error {
	parsedURL, err := url.Parse(repo.url)
	if err != nil {
		return CliErrorf(err, "invalid URL %q", repo.url)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return CliErrorf(nil, "invalid GitLab URL format (expected owner/repo or group/repo) found %q", repo.url)
	}
	projectPath := strings.Join(pathParts, "/")

	client, err := getGitlabClient()
	if err != nil {
		return CliErrorf(err, "failed to get GitLab client for %q", repo.url)
	}

	ctx := context.Background()

	glProject, _, err := client.Projects.GetProject(projectPath, nil, gitlab.WithContext(ctx))
	if err != nil {
		return CliErrorf(err, "failed to fetch GitLab project details for %q", repo.url)
	}

	repo.stars = glProject.StarCount

	defaultBranch := glProject.DefaultBranch

	listCommitOptions := &gitlab.ListCommitsOptions{
		RefName: gitlab.Ptr(defaultBranch),
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
			Page:    1,
		},
	}

	commits, _, err := client.Commits.ListCommits(projectPath, listCommitOptions, gitlab.WithContext(ctx))
	if err != nil {
		return CliErrorf(err, "failed to fetch latest commit for default branch %q on %q", defaultBranch, repo.url)
	}

	if len(commits) > 0 {
		latestCommit := commits[0]
		repo.lastUpdate = *latestCommit.CommittedDate
	}

	return nil
}

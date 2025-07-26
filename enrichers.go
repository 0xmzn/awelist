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
	"github.com/gosimple/slug"
)

type Enricher interface {
	Enrich() error
}

type RemoteRepo interface {
	Stars() int
	LastUpdate() time.Time
	Enrich() error
}

type Slugifier struct {
	original string
	slug     string
}

type githubClient struct {
	apiToken string
	client   *github.Client
	once     sync.Once
	initErr  error
}

var ghClientSingleton = &githubClient{}

func getGitHubClient() (*github.Client, error) {
	ghClientSingleton.once.Do(func() {
		token := os.Getenv("GITHUB_API_KEY")
		if token == "" {
			ghClientSingleton.initErr = CliErrorf(nil, "GITHUB_API_KEY environment variable is not set. A token is required for GitHub API calls.")
			return
		}
		ghClientSingleton.apiToken = token
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

// TODO: this will fail if it's anything but a provider.com/user/repo. e.g: https://github.com/trending?l=go
func NewRemoteRepo(url string) RemoteRepo {
	if strings.HasPrefix(url, "https://github.com") {
		return &githubRepo{url: url}
	}
	if strings.HasPrefix(url, "https://gitlab.com") {
		return &gitlabRepo{url: url}
	}
	return nil
}

func NewSlugifier(str string) *Slugifier {
	return &Slugifier{original: str}
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

	// TODO: Wrong feild. Figure out how to get the data of the latest commit on default branch.
	// updated_at will be updated any time the repository object is updated,
	// e.g. when the description or the primary language of the repository is updated.
	// stackoverflow: https://stackoverflow.com/questions/15918588/github-api-v3-what-is-the-difference-between-pushed-at-and-updated-at
	if ghRepo.UpdatedAt != nil {
		repo.lastUpdate = ghRepo.UpdatedAt.Time
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
	repo.stars = 400
	return nil
}

func (s *Slugifier) Slug() string {
	return s.slug
}

func (s *Slugifier) Enrich() error {
	s.slug = slug.Make(s.original)
	return nil
}

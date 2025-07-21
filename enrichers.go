package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/github"
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
	slug string
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

	ctx := context.Background()
	httpClient := http.DefaultClient

	client := github.NewClient(httpClient)

	ghRepo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return CliErrorf(err, "failed to fetch github repo details %q", repo.url)
	}

	if ghRepo.StargazersCount != nil {
		repo.stars = *ghRepo.StargazersCount
	}

	if ghRepo.UpdatedAt != nil {
		repo.lastUpdate = ghRepo.UpdatedAt.Time
	}
	fmt.Println("Here")

	fmt.Println(ghRepo.String())

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

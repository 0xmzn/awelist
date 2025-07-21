package main

import (
	"context"
	"github.com/google/go-github/github"
	"net/http"
	"net/url"
	"strings"
)

type Enricher interface {
	Enrich() error
}

type RemoteRepo interface {
	Stars() int
	Enrich() error
}

type githubRepo struct {
	url   string
	stars int
}

type gitlabRepo struct {
	url   string
	stars int
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

func (repo *githubRepo) Stars() int {
	return repo.stars
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
	var httpClient *http.Client
	httpClient = http.DefaultClient

	client := github.NewClient(httpClient)

	ghRepo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return CliErrorf(err, "failed to fetch github repo details %q", repo.url)
	}

	if ghRepo.StargazersCount != nil {
		repo.stars = *ghRepo.StargazersCount
	}

	return nil
}

func (repo *gitlabRepo) Stars() int {
	return repo.stars
}

func (repo *gitlabRepo) Enrich() error {
	repo.stars = 400
	return nil
}

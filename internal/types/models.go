package types

import "time"

type GitRepoMetadata struct {
	Stars      int       `json:"stars"`
	LastUpdate time.Time `json:"last_update"`
}

type Link struct {
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
	URL         string `yaml:"url" json:"url"`

	RepoMetadata *GitRepoMetadata `yaml:"-" json:"repo_metadata,omitempty"`
}

type Category struct {
	Title         string      `yaml:"title" json:"title"`
	Description   string      `yaml:"description,omitempty" json:"description,omitempty"`
	Slug          string      `yaml:"-" json:"slug"`
	Links         []*Link     `yaml:"links,omitempty" json:"links,omitempty"`
	Subcategories []*Category `yaml:"subcategories,omitempty" json:"subcategories,omitempty"`
}

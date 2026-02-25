package types

import "time"

type GitRepoMetadata struct {
	Stars      int       `json:"stars"`
	IsArchived bool      `json:"is_archived"`
	LastUpdate time.Time `json:"last_update"`
	EnrichedAt time.Time `json:"enriched_at"`
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

type AwesomeList []*Category

type LockFile struct {
	Metadata LockMetadata `json:"metadata"`
	List     AwesomeList  `json:"list"`
}

type ProviderMetrics struct {
	Provider   string `json:"provider"`
	Attempted  int    `json:"attempted"`
	Successful int    `json:"successful"`
	Failed     int    `json:"failed"`
}

type LockMetadata struct {
	UpdatedAt       time.Time         `json:"updated_at"`
	ProviderMetrics []ProviderMetrics `json:"provider_metrics,omitempty"`
	FailedLinks     map[string]string `json:"failed_links,omitempty"`
}

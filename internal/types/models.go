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

type AwesomeList []*Category

func (l AwesomeList) TotalCount() int {
	count := 0
	for _, c := range l {
		count += c.countRecursive()
	}
	return count
}

func (l AwesomeList) Flatten() []*Link {
	var links []*Link
	for _, c := range l {
		links = append(links, c.collectRecursive()...)
	}
	return links
}

func (c *Category) countRecursive() int {
	n := len(c.Links)
	for _, sub := range c.Subcategories {
		n += sub.countRecursive()
	}
	return n
}

func (c *Category) collectRecursive() []*Link {
	all := c.Links
	for _, sub := range c.Subcategories {
		all = append(all, sub.collectRecursive()...)
	}
	return all
}

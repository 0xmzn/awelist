package main

import "time"

type BaseCategory struct {
	Title         string         `yaml:"title"`
	Description   string         `yaml:"description,omitempty"`
	Links         []BaseLink     `yaml:"links"`
	Subcategories []BaseCategory `yaml:"subcategories,omitempty"`
}

type BaseLink struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Url         string `yaml:"url"`
}

type baseAwesomelist []BaseCategory

type EnrichedCategory struct {
	Title         string             `json:"title"`
	Description   string             `json:"description,omitempty"`
	Links         []EnrichedLink     `json:"links"`
	Subcategories []EnrichedCategory `json:"subcategories,omitempty"`

	Slug string `json:"slug"`
}

type EnrichedLink struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`

	IsRepo     bool      `json:"isRepo"`
	Stars      int       `json:"stars"`
	LastUpdate time.Time `json:"lastUpdate"`
	IsArchived bool      `json:"isArchived"`
}

type enrichedAwesomelist []EnrichedCategory

package main

type Category struct {
	Title         string     `yaml:"title"`
	Description   string     `yaml:"description,omitempty"`
	Link          []Link     `yaml:"links"`
	Subcategories []Category `yaml:"subcategories,omitempty"`
}

type Link struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Url         string `yaml:"url"`

	// derived
	Stars int `yaml:"-"`
}

type awesomeList []Category

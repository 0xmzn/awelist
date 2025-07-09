package main

type Category struct {
	Title         string     `yaml:"title"`
	Description   string     `yaml:"description"`
	Records       []Record   `yaml:"records"`
	Subcategories []Category `yaml:"subcategories,omitempty"`
}

type Record struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	RecordData  string `yaml:"recordData"`
}

type awesomeList []Category

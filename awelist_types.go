package main

type Category struct {
	Title         string     `yaml:"title"`
	Description   string     `yaml:"description,omitempty"`
	Records       []Record   `yaml:"records"`
	Subcategories []Category `yaml:"subcategories,omitempty"`
}

type Record struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	RecordData  string `yaml:"recordData,omitempty"`
}

type awesomeList []Category

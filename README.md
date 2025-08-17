# `awelist`: A CLI Tool for Managing Awesome Lists

`awelist` is a lightweight command-line tool written in Go that helps you automate the maintenance and publishing of "awesome lists." It streamlines the process of keeping your lists up-to-date by fetching metadata and generating a consistent, well-formatted output file.

-----

## Features

  * **Structured List Management**: Define your awesome list content in a structured **YAML file** (`awesome.yaml`).
  * **Automatic Enrichment**: Automatically fetch and add metadata like **GitHub/GitLab stars**, `last commit date`, and archived status for each listed project.
  * **Template-Based Generation**: Use custom go templates to generate various output formats, such as `README.md`, HTML, or whatever you feel like.
  * **Easy Contributions**: Add new links or categories to your list directly from the command line without manually editing the YAML file.

-----

## Installation

1. **Clone the repository:**

```bash
git clone https://github.com/0xmzn/awelist.git
cd awelist
```

2. **Build the tool:**

```bash
go build .
```


-----

## Usage

The tool has three main commands: `add`, `enrich`, and `generate`.

### Add a new link or category

Use the `add` command to quickly append new items to your list.

#### Adding a link

```bash
awelist add link --title "Awelist" --description "A cool project." --url "https://github.com/user/repo" "Some Sub Category"
```

*The `"Some Sub Category"` part specifies the path where the new link will be added. You don't have to worry about the exact name and formatting of the sub category (e.g. uppercase vs lowercase), awelist will manage it for you. Just make sure to write the exact name of the category your adding to.*

#### Adding a category

```bash
awelist add category --title "Toyota" --description "A reliable car." "vehicles"
```

if you're trying to add a category to the top-level list, you can use the special `.` argument:

```bash
awelist add category --title "Vehicles" --description "Things that go vroom" .
```

### Enrich the list with metadata

The `enrich` command fetches data (like stars and last update dates) and saves an enriched version of your list in a new file, `awesome-lock.json`.

```bash
awelist enrich
```
By default, it uses `awesome.yaml` file in the current directory. You can specify which file to load data from using `--awesome-file,-f` which is a global flag.

## API Keys:

Awelist reads GitHub & GitLab API keys from environment variables to fetch repositories' metadata. You're only required provide API key for the provider you're using.

- GitHub: `GITHUB_API_KEY`
- GitLab: `GILAB_API_KEY`

### Generate a new `README.md`

Use the `generate` command with a template file to create your final output.

```bash
awelist generate my-template.md > README.md
```

-----

## Example `awesome.yaml`

You can take a look at [awesome.yaml](awesome.yaml) for example. It contains an `AwesomeList` whose structure follows this:

```go
type Category struct {
	Title         string
	Description   string      // can be omitted
	Links         []Link
	Subcategories []Category  // can be omitted
}

type Link struct {
	Title       string
	Description string
	Url         string
}

type AwesomeList []Category
```
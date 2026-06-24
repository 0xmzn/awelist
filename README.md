# `awelist`: A CLI Tool for Managing Awesome Lists

`awelist` is a Go CLI that automates maintaining and publishing "[Awesome Lists](https://github.com/sindresorhus/awesome)." It fetches metadata for listed projects and generates a formatted output file.

The initial proposal that led to this tool is [here](https://github.com/avelino/awesome-go/issues/5662).

-----

## Features

  * Define your list content in a structured YAML file (`awesome.yaml`).
  * Fetch metadata like GitHub stars, archived status, and last commit date for each project.
  * Use Go templates to generate whatever output you need.
  * Add new links or categories from the command line without editing YAML by hand.

-----

## Installation

### Using `go install`

```bash
go install github.com/0xmzn/awelist/cmd/awelist@latest
```

### Build from source

1. Clone the repository:

```bash
git clone https://github.com/0xmzn/awelist.git
cd awelist
```

2. Build the tool:

```bash
go mod tidy
go build ./cmd/awelist
```

-----

## Usage

The tool has four main commands: `add`, `enrich`, `generate` and `report`.

```
Usage: awelist <command> [flags]

A CLI tool for managing awesome lists

Flags:
  -h, --help                   Show context-sensitive help.
      --file="awesome.yaml"    Path to awesome.yaml.
      --lock="awesome-lock.json"
                               Path to awesome-lock.json.

Commands:
  add link        Add a new link to a category.
  add category    Add a new subcategory to a category.
  enrich          Enrich YAML file. on success, awesome-lock.json file will be
                  created.
  generate        Generate file from template. uses dry data if
                  awesome-lock.json does not exist.
  report          Show details from the last enrichment.

Run "awelist <command> --help" for more information on a command.
```

### Add a new link or category

The `add` command appends new items to your list.

#### Adding a link

```bash
awelist add link --title "Zap" --description "Blazing fast, structured, leveled logging in Go." --url "github.com/uber-go/zap" "Logging"
```

`"Logging"` is the target category. Case and formatting don't matter (awelist normalizes them), but the category itself needs to exist in your list.

#### Adding a category

```bash
awelist add category --title "Logging" --description "All things logging" "Utilites"
```

To add a category to the top-level list, use `.` as the path:

```bash
awelist add category --title "Logging" --description "All things logging" .
```

### Enrich the list with metadata

The `enrich` command fetches metadata for each link and writes the results to `awesome-lock.json`. See [awesome-lock.json](awesome-lock.json) for an example.

```bash
awelist enrich
```
By default, awelist reads from `awesome.yaml` in the current directory. Use the `--file` global flag to specify a different file.

Cached enrichments go stale after 24 hours. Change this with `--ttl`:

```bash
awelist enrich --ttl 48h
```

#### Supported providers

Awelist supports two providers:

| Provider | Detected URLs | Metadata Fetched |
| :--- | :--- | :--- |
| **GitHub** | `github.com/*` | Stars, archived status, last commit date |
| **GitLab** | `gitlab.com/*` | Stars, archived status, last commit date |

Links that don't match a supported provider are skipped and reported as unhandled.

#### API keys

Awelist reads API keys from environment variables. You only need a key for the provider you're using.

- GitHub: `GITHUB_TOKEN`
- GitLab: `GITLAB_TOKEN`

### Generate a new `README.md`

The `generate` command takes a template file and produces your output. It uses `awesome-lock.json` as its data source. If that file doesn't exist, awelist builds the data on the fly without making any remote calls.
```bash
awelist generate my-template.md > README.md
```

See [templates/readme.template](templates/readme.template) for an example template.

-----

## Data structure for templating

### 1. Structure of `awesome.yaml` (the input)

[awesome.yaml](./awesome.yaml) is the source of truth for awelist. It's an array of Category objects.

#### Category fields (YAML input)

| Field | Type | Description |
| :--- | :--- | :--- |
| **Title** | `string` | The name of the category. |
| **Description** | `string` | An optional, brief description. |
| **Links** | `List of Link` | An optional list of projects. |
| **Subcategories** | `List of Category` | An optional list of nested categories. |

#### Link fields (YAML input)

| Field | Type | Description |
| :--- | :--- | :--- |
| **Title** | `string` | The name of the project. |
| **Description** | `string` | A short description. |
| **URL** | `string` | The project's URL. |

### 2. Enriched data for templates (`awesome-lock.json`)

Templates run against the enriched data from `awesome-lock.json`. The structure is the same as the YAML input, with extra fields on Category and Link.

#### Enriched category fields
| Field | Description | Go Template Access |
| :--- | :--- | :--- |
| `Slug` | A URL-friendly version of the Title (e.g., `logging`). | `{{ .Slug }}` |
| *(All other Category fields are also available)* | | | |

#### Enriched link fields
Each link gets a `RepoMetadata` object with these fields:

| Field | Description | Go Template Access |
| :--- | :--- | :--- |
| `RepoMetadata.Stars` | The number of stars (e.g., GitHub stargazers). | `{{ .RepoMetadata.Stars }}` |
| `RepoMetadata.IsArchived` | Whether the repository is archived. | `{{ .RepoMetadata.IsArchived }}` |
| `RepoMetadata.LastUpdate` | The date of the last commit on the default branch. | `{{ .RepoMetadata.LastUpdate }}` |
| `RepoMetadata.EnrichedAt` | Last time link was enriched. | `{{ .RepoMetadata.EnrichedAt }}` |
| *(All other Link fields are also available)* | | |

### Example

#### Input YAML
```yaml
- title: Command Line
  subcategories:
  - title: Advanced Console UIs
    description: Libraries for building Console Applications and Console User Interfaces.
    links:
    - title: bubbles
      description: TUI components for bubbletea.
      url: https://github.com/charmbracelet/bubbles
    - title: bubbletea
      description: Go framework to build terminal apps, based on The Elm Architecture.
      url: https://github.com/charmbracelet/bubbletea
  - title: Standard CLI
    description: Libraries for building standard or basic Command Line applications.
    links:
    - title: cobra
      description: Commander for modern Go CLI interactions.
      url: https://github.com/spf13/cobra
    - title: pflag
      description: Drop-in replacement for Go's flag package, implementing POSIX/GNU-style --flags.
      url: https://github.com/spf13/pflag
```

#### Generated JSON
```json
[
  {
    "title": "Command Line",
    "slug": "command-line",
    "subcategories": [
      {
        "title": "Advanced Console UIs",
        "description": "Libraries for building Console Applications and Console User Interfaces.",
        "slug": "advanced-console-uis",
        "links": [
          {
            "title": "bubbles",
            "description": "TUI components for bubbletea.",
            "url": "https://github.com/charmbracelet/bubbles",
            "repo_metadata": {
              "stars": 8570,
              "is_archived": false,
              "last_update": "2026-06-15T09:07:40Z",
              "enriched_at": "2026-06-21T18:03:51.961707+02:00"
            }
          },
          {
            "title": "bubbletea",
            "description": "Go framework to build terminal apps, based on The Elm Architecture.",
            "url": "https://github.com/charmbracelet/bubbletea",
            "repo_metadata": {
              "stars": 43270,
              "is_archived": false,
              "last_update": "2026-06-01T16:34:19Z",
              "enriched_at": "2026-06-21T18:03:51.961709+02:00"
            }
          }
        ]
      },
      {
        "title": "Standard CLI",
        "description": "Libraries for building standard or basic Command Line applications.",
        "slug": "standard-cli",
        "links": [
          {
            "title": "cobra",
            "description": "Commander for modern Go CLI interactions.",
            "url": "https://github.com/spf13/cobra",
            "repo_metadata": {
              "stars": 44137,
              "is_archived": false,
              "last_update": "2026-04-25T23:07:41Z",
              "enriched_at": "2026-06-21T18:03:51.96171+02:00"
            }
          },
          {
            "title": "pflag",
            "description": "Drop-in replacement for Go's flag package, implementing POSIX/GNU-style --flags.",
            "url": "https://github.com/spf13/pflag",
            "repo_metadata": {
              "stars": 2743,
              "is_archived": false,
              "last_update": "2026-06-06T14:20:53Z",
              "enriched_at": "2026-06-21T18:03:51.961711+02:00"
            }
          }
        ]
      }
    ]
  }
]
```

# `awelist`: A CLI Tool for Managing Awesome Lists

**WORK IN PROGRESS**
---

`awelist` is a command-line tool written in Go that helps automate the maintenance and publishing of "[Awesome Lists](https://github.com/sindresorhus/awesome)." It streamlines the process of keeping your lists up-to-date by fetching metadata and generating a consistent, well-formatted output file.

The initial proposal that lead to the birth of this tool can be found [here](https://github.com/avelino/awesome-go/issues/5662).  

-----

## Features

  * **Structured List Management**: Define your awesome list content in a structured **YAML file** (`awesome.yaml`).
  * **Automatic Enrichment**: Automatically fetch and add metadata like **GitHub/GitLab stars** and archived status for each listed project.
  * **Template-Based Generation**: Use custom go templates to generate various output formats, such as `README.md`, HTML, or whatever you feel like.
  * **Easy Contributions**: Add new links or categories to your list directly from the command line without manually editing the YAML file.

-----

## Installation

### Build from source

1. **Clone the repository:**

```bash
git clone https://github.com/0xmzn/awelist.git
cd awelist
```

2. **Build the tool:**

```bash
go mod tidy
go build ./cmd/awelist
```

-----

## Usage

The tool has three main commands: `add`, `enrich`, and `generate`.

### Add a new link or category

Use the `add` command to quickly append new items to your list.

#### Adding a link

```bash
awelist add link --title "Zap" --description "Blazing fast, structured, leveled logging in Go." --url "github.com/uber-go/zap" "Logging"
```

*The `"Logging"` part specifies the path where the new link will be added. You don't have to worry about the exact name and formatting of the sub category (e.g. uppercase vs lowercase), awelist will manage it for you. Make sure to write the exact name of the category you're adding to.*

#### Adding a category

```bash
awelist add category --title "Logging" --description "All things logging" "Utilites"
```

if you're trying to add a category to the top-level list, you can use the special `.` argument:

```bash
awelist add category --title "Logging" --description "All things logging" .
```

### Enrich the list with metadata

The `enrich` command fetches data (like stars and last update dates) and saves an enriched version of your list in a new file, `awesome-lock.json`. Take a look at [awesome-lock.json](awesome-lock.json) for example.

```bash
awelist enrich
```
By default, awelist uses `awesome.yaml` file in the current directory. You can specify which file to load data from using `--awesome-file,-f` which is a global flag.

#### API Keys:

Awelist reads GitHub & GitLab API keys from environment variables to fetch repositories' metadata. You're only required provide API key for the provider you're using.

- GitHub: `GITHUB_TOKEN`
- GitLab: `GITLAB_TOKEN`

### Generate a new `README.md`

Use the `generate` command with a template file to create your final output. By default, generate relies on `awesome-lock.json` to generate template files. If the file doesn't exist, awelist will generate an on-the-fly enriched list without any remote calls.
```bash
awelist generate my-template.md > README.md
```

Example template can be found under [templates/readme.template](templates/readme.template).

-----

## Data Structure for Templating

### 1. Structure of `awesome.yaml` (The Input)

[awesome.yaml](./awesome.yaml) is the source-of-truth of data used by awelist. It is an array of **Category** objects.

#### Category Fields (YAML Input)

| Field | Type | Description |
| :--- | :--- | :--- |
| **Title** | `string` | The name of the category. |
| **Description** | `string` | An optional, brief description. |
| **Links** | `List of Link` | An optional list of projects. |
| **Subcategories** | `List of Category` | An optional list of nested categories. |

#### Link Fields (YAML Input)

| Field | Type | Description |
| :--- | :--- | :--- |
| **Title** | `string` | The name of the project. |
| **Description** | `string` | A short description. |
| **URL** | `string` | The project's URL. |

### 2. Enriched Data for Templates (`awesome-lock.json`)

Templates are executed against the enriched data loaded from `awesome-lock.json`. This structure is identical to the YAML input, but with new fields added to `Category` and `Link`.

#### Enriched Category Fields
| Field | Description | Go Template Access |
| :--- | :--- | :--- |
| `Slug` | A URL-friendly version of the Title (e.g., `logging`). | `{{ .Slug }}` |
| *(All other Category fields are also available)* | | | |

#### Enriched Link Fields
Links gain a `RepoMetadata` object with the following fields:

| Field | Description | Go Template Access |
| :--- | :--- | :--- |
| `RepoMetadata.Stars` | The number of stars (e.g., GitHub stargazers). | `{{ .RepoMetadata.Stars }}` |
| `RepoMetadata.IsArchived` | Whether the repository is archived. | `{{ .RepoMetadata.IsArchived }}` |
| `RepoMetadata.EnrichedAt` | Last time link was enriched | `{{ .RepoMetadata.EnrichedAt }}` |
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
              "stars": 7688,
              "is_archived": false,
              "enriched_at": "2026-02-05T11:34:36.210692848+01:00"
            }
          },
          {
            "title": "bubbletea",
            "description": "Go framework to build terminal apps, based on The Elm Architecture.",
            "url": "https://github.com/charmbracelet/bubbletea",
            "repo_metadata": {
              "stars": 39085,
              "is_archived": false,
              "enriched_at": "2026-02-05T11:34:36.619674976+01:00"
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
              "stars": 43083,
              "is_archived": false,
              "enriched_at": "2026-02-05T11:34:37.029426395+01:00"
            }
          },
          {
            "title": "pflag",
            "description": "Drop-in replacement for Go's flag package, implementing POSIX/GNU-style --flags.",
            "url": "https://github.com/spf13/pflag",
            "repo_metadata": {
              "stars": 2695,
              "is_archived": false,
              "enriched_at": "2026-02-05T11:34:37.439142461+01:00"
            }
          }
        ]
      }
    ]
  }
]
```

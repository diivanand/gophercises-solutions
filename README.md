# Gophercises Solutions

My solutions to the [Gophercises](https://gophercises.com) Go exercises by Jon Calhoun.

## Progress

| #  | Exercise                                     | Directory      | Status         |
|----|----------------------------------------------|----------------|----------------|
| 1  | Quiz Game                                    | `quiz`         | ✅ Complete    |
| 2  | URL Shortener                                | `urlshort`     | ✅ Complete    |
| 3  | Choose Your Own Adventure                    | `cyoa`         | 🚧 Not started |
| 4  | HTML Link Parser                             | `link`         | 🚧 Not started |
| 5  | Sitemap Builder                              | `sitemap`      | 🚧 Not started |
| 6  | Hacker Rank Problems &mdash; strings & bytes | `hr1`          | 🚧 Not started |
| 7  | CLI Task Manager                             | `task`         | 🚧 Not started |
| 8  | Phone Number Normalizer                      | `phone`        | 🚧 Not started |
| 9  | Deck of Cards                                | `deck`         | 🚧 Not started |
| 10 | Blackjack Game                               | `blackjack`    | 🚧 Not started |
| 11 | Blackjack AI                                 | `blackjack_ai` | 🚧 Not started |
| 12 | File Renaming Tool                           | `renamer`      | 🚧 Not started |
| 13 | Quiet HN                                     | `quiet_hn`     | 🚧 Not started |

### Notes on individual exercises

- **`urlshort`** implements `MapHandler`, `YAMLHandler`, and the JSON bonus as
  `JSONHandler`, all covered by unit tests. The Bolt-database bonus is not done.
  `main.go` adds two flags beyond upstream:

  ```sh
  cd urlshort
  go run .                            # built-in YAML example (the default)
  go run . -format=json               # built-in JSON example, same paths
  go run . -in paths.yaml             # read the mappings from a file instead
  go run . -format=json -in paths.json
  ```

  Unmatched paths fall through a chain: the chosen format handler, then a
  hardcoded `MapHandler`, then a `ServeMux` that serves `Hello, world!`.

  YAML parsing uses [`go.yaml.in/yaml/v3`](https://github.com/yaml/go-yaml)
  rather than upstream's `gopkg.in/yaml.v2`. The original `go-yaml/yaml`
  repository was archived in April 2025 and its v1&ndash;v3 are now frozen at
  security fixes only; the YAML org's fork is drop-in compatible and only the
  import path changed. JSON uses the standard library's `encoding/json`, so the
  bonus adds no dependency.

- **`quiet_hn`** has a working upstream Hacker News client in `lib/`, with tests.
  Those tests call the live HN API, so they fail without a network connection.
  The exercise itself &mdash; making the page load quietly and quickly &mdash; is
  not done.

The not-started exercises contain an empty `main()` and an empty `lib` package.

## Layout

Each exercise is its own Go module, named after its directory:

```
<exercise>/
  go.mod          module <exercise>, go 1.26.4
  .gitignore
  README.md       upstream exercise instructions
  main.go         package main
  lib/
    <exercise>.go package lib
```

Implementation lives in `lib` so it can be unit tested, with `main.go` reduced to
flag parsing and wiring. Some exercises also keep upstream asset files, such as
`cyoa/gopher.json`, `link/ex1.html`, `renamer/sample/`, and `quiet_hn/index.gohtml`.

There is no `go.work` tying the modules together, so `go build ./...` from the repo
root fails with `directory prefix . does not contain main module`. Work from inside
an exercise directory:

```sh
cd quiz
go build ./...
go test ./...
go run . -h
```

## License

See [LICENSE](LICENSE).

# Gophercises Solutions

My solutions to the [Gophercises](https://gophercises.com) Go exercises by Jon Calhoun.

## Progress

| #  | Exercise                                     | Directory      | Status         |
|----|----------------------------------------------|----------------|----------------|
| 1  | Quiz Game                                    | `quiz`         | ✅ Complete    |
| 2  | URL Shortener                                | `urlshort`     | 🚧 Not started |
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

### Notes on the not-started exercises

`urlshort` and `quiet_hn` ship upstream code and are further along than the bare
scaffolds:

- **`urlshort`** has upstream's `MapHandler` and `YAMLHandler` stubs, both still
  `// TODO: Implement this...` returning `nil`. The server starts, but every
  request panics on the nil `http.HandlerFunc` until they're implemented.
- **`quiet_hn`** has a working upstream Hacker News client in `lib/`, with tests.
  Those tests call the live HN API, so they fail without a network connection.
  The exercise itself &mdash; making the page load quietly and quickly &mdash; is
  not done.

The remaining exercises contain an empty `main()` and an empty `lib` package.

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

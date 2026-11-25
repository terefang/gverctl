<!-- README.md for the gverctl Go CLI tool -->

# gverctl

* Command-line tool written in Go for reading, modifying, and writing version strings.
* Matches version strings in the format `R.M.N` with an optional pre-release or support tag (`-alpha`, `-beta`, `-rc`, or `-lts`).
* Reads input from a file or `stdin` and writes output to a file or `stdout`.

## Features

* Prints the CLI tool version with `-version`.
* Sets individual version components (`R`, `M`, `N`, and tag) or clears the tag.
* Increments individual numeric components (`R`, `M`, `N`) by `1`.
* Supports file I/O (`-in`, `-out`) and stream I/O (`stdin`, `stdout`).

## Requirements

* Go installed on your system.
* A `main.go` source file containing the `gverctl` implementation.

## Build and Run

* Run directly with Go:
    * `go run main.go -version`
* Build a binary:
    * `go build -o gverctl main.go`
* Run the built binary:
    * `./gverctl -version`

## Flags

* `-version`: Prints the CLI tool version (`0.1.0`) to `stdout` and exits.
* `-in <path>`: Reads the input version string from `<path>`; reads from `stdin` if omitted.
* `-out <path>`: Writes the resulting version string to `<path>`; writes to `stdout` if omitted.
* `-set-r <int>`: Sets `R` to a non-negative integer.
* `-set-m <int>`: Sets `M` to a non-negative integer.
* `-set-n <int>`: Sets `N` to a non-negative integer.
* `-set-tag <tag>`: Sets the tag to `alpha`, `beta`, `rc`, or `lts`.
* `-clear-tag`: Removes the tag component.
* `-inc-r`: Increments `R` by `1`.
* `-inc-m`: Increments `M` by `1`.
* `-inc-n`: Increments `N` by `1`.

## Usage Examples

* Print the tool version:
    * `go run main.go -version`
* Read from `stdin`, increment `N`, and write to `stdout`:
    * `echo "1.2.3-rc" | go run main.go -inc-n`
* Read from a file, set `M` and `tag`, and write to another file:
    * `go run main.go -in version.txt -out next_version.txt -set-m 5 -set-tag lts`
* Read from `stdin` and clear the tag:
    * `echo "2.0.0-alpha" | go run main.go -clear-tag`

## Code Structure Explanation

* `toolVersion`: Constant holding the CLI tool version string (`"0.1.0"`).
* `versionPattern`: Regular expression matching `^(\d+)\.(\d+)\.(\d+)(?:-(alpha|beta|rc|lts))?$`.
* `Version`: Struct storing `R`, `M`, `N`, `Tag`, and `HasTag`.
* `ParseVersion`: Trims input whitespace, validates the format, parses numeric components, and extracts the optional tag.
* `(v Version) String`: Formats `R.M.N` and appends `-<Tag>` when `HasTag` is true and `Tag` is non-empty.
* `validateTag`: Checks that a tag is empty or one of `alpha`, `beta`, `rc`, or `lts`.
* `readInput`: Reads raw text from `-in` if given, otherwise from the `io.Reader` (`stdin`).
* `writeOutput`: Appends `\n` and writes to `-out` if given, otherwise to the `io.Writer` (`stdout`).
* `run`: Creates the `gverctl` `flag.FlagSet`, parses flags, handles `-version`, reads and parses input, applies set/clear/increment operations, and writes output.
* `main`: Calls `run` with `os.Args[1:]`, `os.Stdin`, and `os.Stdout`, printing errors to `os.Stderr` and exiting with status `1` on failure.
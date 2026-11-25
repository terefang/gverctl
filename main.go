// Package main implements a command-line tool for reading, modifying, and writing
// version strings formatted as R.M.N(-alpha|-beta|-rc|-lts).
package main

import (
    // bufio provides buffered I/O for reading from standard input.
    "bufio"
    // pkg provides shared package constants such as PkgVersion.
    "gversion/pkg"

    // errors provides error creation and inspection helpers.
    "errors"
    // flag provides command-line flag parsing.
    "flag"
    // fmt provides formatted I/O functions.
    "fmt"
    // io provides basic I/O interfaces.
    "io"
    // os provides operating system functionality such as file and stream access.
    "os"
    // regexp provides regular expression matching.
    "regexp"
    // strconv provides conversions between strings and basic data types.
    "strconv"
    // strings provides string manipulation utilities.
    "strings"
    // time provides current date and time access for -drel.
    "time"
)

// toolVersion holds the version string of the CLI tool itself.
const toolVersion = pkg.PkgVersion

// versionPattern matches R.M.N followed by an optional pre-release tag
// (-alpha, -beta, -rc, or -lts) and an optional -DD suffix.
var versionPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(alpha|beta|rc|lts))?$`)

// Version holds the parsed components of a version string.
type Version struct {
    // R is the first numeric component.
    R int
    // M is the second numeric component.
    M int
    // N is the third numeric component.
    N int
    // Tag is the optional pre-release or support label (alpha, beta, rc, lts).
    Tag string
    // HasTag records whether a tag was present or explicitly set.
    HasTag bool
}

// ParseVersion parses a version string in the form R.M.N(-alpha|-beta|-rc|-lts)(-DD).
func ParseVersion(raw string) (Version, error) {
    // Trim surrounding whitespace from the input version string.
    trimmed := strings.TrimSpace(raw)
    // Match the trimmed string against versionPattern.
    matches := versionPattern.FindStringSubmatch(trimmed)
    // Return an error if the string does not match the required format.
    if matches == nil {
        return Version{}, fmt.Errorf("invalid version string %q: expected R.M.N(-alpha|-beta|-rc|-lts)", trimmed)
    }

    // Parse R as a base-10 integer.
    rVal, err := strconv.Atoi(matches[1])
    if err != nil {
        return Version{}, fmt.Errorf("invalid R value %q: %w", matches[1], err)
    }

    // Parse M as a base-10 integer.
    mVal, err := strconv.Atoi(matches[2])
    if err != nil {
        return Version{}, fmt.Errorf("invalid M value %q: %w", matches[2], err)
    }

    // Parse N as a base-10 integer.
    nVal, err := strconv.Atoi(matches[3])
    if err != nil {
        return Version{}, fmt.Errorf("invalid N value %q: %w", matches[3], err)
    }

    // Construct the Version struct with core numeric fields.
    v := Version{
        R: rVal,
        M: mVal,
        N: nVal,
    }

    // Populate optional tag if present.
    if matches[4] != "" {
        v.Tag = matches[4]
        v.HasTag = true
    }

    // Return the parsed Version struct.
    return v, nil
}

// String formats the Version struct back into R.M.N(-alpha|-beta|-rc|-lts)(-DD).
func (v Version) String() string {
    // Start with the mandatory R.M.N core.
    var b strings.Builder
    fmt.Fprintf(&b, "%d.%d.%d", v.R, v.M, v.N)

    // Append the optional tag if set.
    if v.HasTag && v.Tag != "" {
        b.WriteString("-")
        b.WriteString(v.Tag)
    }

    // Return the formatted version string.
    return b.String()
}

// validateTag checks that a tag is one of the allowed values or empty.
func validateTag(tag string) error {
    switch tag {
    case "", "alpha", "beta", "rc", "lts":
        return nil
    default:
        return fmt.Errorf("invalid tag %q: allowed values are alpha, beta, rc, lts", tag)
    }
}

// readInput reads the raw version string from inputPath if non-empty, or from r otherwise.
func readInput(inputPath string, r io.Reader) (string, error) {
    // Read from file when -in is provided.
    if inputPath != "" {
        data, err := os.ReadFile(inputPath)
        if err != nil {
            return "", fmt.Errorf("failed to read input file %q: %w", inputPath, err)
        }
        return string(data), nil
    }

    // Read from the provided reader (such as os.Stdin).
    reader := bufio.NewReader(r)
    data, err := io.ReadAll(reader)
    if err != nil {
        return "", fmt.Errorf("failed to read input stream: %w", err)
    }
    return string(data), nil
}

// writeOutput writes the final version string to outputPath if non-empty, or to w otherwise.
func writeOutput(outputPath string, w io.Writer, output string) error {
    // Append a trailing newline for standard text/CLI output.
    payload := []byte(output + "\n")

    // Write to file when -out is provided.
    if outputPath != "" {
        if err := os.WriteFile(outputPath, payload, 0644); err != nil {
            return fmt.Errorf("failed to write output file %q: %w", outputPath, err)
        }
        return nil
    }

    // Write to the provided writer (such as os.Stdout).
    if _, err := w.Write(payload); err != nil {
        return fmt.Errorf("failed to write output stream: %w", err)
    }
    return nil
}

// run executes the CLI logic with the given arguments, input reader, and output writer.
func run(args []string, stdin io.Reader, stdout io.Writer) error {
    // Create a dedicated FlagSet for argument parsing.
    fs := flag.NewFlagSet("gverctl", flag.ContinueOnError)
    // Direct flag usage output to stdout.
    fs.SetOutput(stdout)

    // Help flags.
    showHelp := fs.Bool("help", false, "print help and usage information and exit")
    showHelpShort := fs.Bool("h", false, "print help and usage information and exit")

    // Tool version flag.
    showVersion := fs.Bool("version", false, "print the CLI tool version and exit")

    // Version string argument flag.
    parseVersionStr := fs.String("parse", "", "version string passed directly as a command-line argument")

    // Input and output flags.
    inFile := fs.String("in", "", "input file path; reads from stdin if omitted")
    outFile := fs.String("out", "", "output file path; writes to stdout if omitted")

    // Flags for setting individual components.
    setR := fs.Int("set-r", -1, "set R component (non-negative integer)")
    setM := fs.Int("set-m", -1, "set M component (non-negative integer)")
    setN := fs.Int("set-n", -1, "set N component (non-negative integer)")
    setTag := fs.String("set-tag", "", "set tag (alpha, beta, rc, lts)")
    clearTag := fs.Bool("clear-tag", false, "remove tag component")

    // Flags for incrementing individual components.
    incR := fs.Bool("inc-r", false, "increment R component by 1")
    incM := fs.Bool("inc-m", false, "increment M component by 1")
    incN := fs.Bool("inc-n", false, "increment N component by 1")

    // High-level release/bump flags.
    drel := fs.Bool("drel", false, "set R to current year, set M to current month, and increment N by 1")
    minor := fs.Bool("minor", false, "increment N by 1")
    major := fs.Bool("major", false, "increment M by 1, N to 0")
    release := fs.Bool("release", false, "increment R by 1, M to 0, N to 0")

    // Parse command-line flags.
    if err := fs.Parse(args); err != nil {
        // Treat explicit help flag errors as clean exits after printing usage.
        if errors.Is(err, flag.ErrHelp) {
            return nil
        }
        return err
    }

    // Print usage/help and exit early when -help or -h is passed.
    if *showHelp || *showHelpShort {
        fs.PrintDefaults()
        return nil
    }

    // Print the tool version and exit early when -version is passed.
    if *showVersion {
        _, err := fmt.Fprintln(stdout, toolVersion)
        return err
    }

    // Obtain raw version input from -parse if given, or from file/stdin otherwise.
    var rawInput string
    var err error
    if *parseVersionStr != "" {
        rawInput = *parseVersionStr
    } else {
        rawInput, err = readInput(*inFile, stdin)
        if err != nil {
            return err
        }
    }

    // Parse input into a Version struct.
    v, err := ParseVersion(rawInput)
    if err != nil {
        return err
    }

    // Apply setting operations for numeric components.
    if *setR >= 0 {
        v.R = *setR
    } else if *setR != -1 {
        return errors.New("set-r must be non-negative")
    }

    if *setM >= 0 {
        v.M = *setM
    } else if *setM != -1 {
        return errors.New("set-m must be non-negative")
    }

    if *setN >= 0 {
        v.N = *setN
    } else if *setN != -1 {
        return errors.New("set-n must be non-negative")
    }

    // Apply tag setting or clearing.
    if *clearTag {
        v.Tag = ""
        v.HasTag = false
    }
    if *setTag != "" {
        if err := validateTag(*setTag); err != nil {
            return err
        }
        v.Tag = *setTag
        v.HasTag = true
    }

    // Apply increment operations for individual components.
    if *incR {
        v.R++
    }
    if *incM {
        v.M++
    }
    if *incN {
        v.N++
    }

    // Apply -drel: set R to current year, M to current month, and increment N by 1.
    if *drel {
        now := time.Now()
        v.R = now.Year()
        v.M = int(now.Month())
        v.N++
    }

    // Apply -minor: increment N by 1.
    if *minor {
        v.N++
    }

    // Apply -major: increment M by 1.
    if *major {
        v.M++
        v.N = 0
    }

    // Apply -release: increment R by 1.
    if *release {
        v.R++
        v.M = 0
        v.N = 0
    }

    // Write the resulting version string to file or stdout.
    return writeOutput(*outFile, stdout, v.String())
}

// main is the entry point of the CLI application.
func main() {
    if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

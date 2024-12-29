![Clank Logo](./clank-preview-image.png)

Clank is a command-line tool written in Go that generates all possible combinations of phone numbers based on a pattern where some digits are unknown. It's particularly useful for generating number sequences when you remember only part of a phone number.

## Features

- Generate all possible combinations for partial phone numbers
- Support for any number of unknown digits using 'x' as placeholders
- Simple command-line interface
- Input validation
- Flexible input methods (with or without flags)

## Installation

### Prerequisites

- Go 1.16 or higher

### Steps

1. Clone the repository:

```bash
git clone https://github.com/yourusername/clank.git
cd clank
```

2. Build the project:

```bash
go build
```

This will create an executable named `clank` (or `clank.exe` on Windows).

## Usage

You can run Clank in two ways:

1. Using the `-n` flag for pattern:

```bash
clank -n 918115605xxx
```

2. Direct pattern input:

```bash
clank 918115605xxx
```

### Flags

- `-n`: Specify the phone number pattern

### Pattern Format

- Use numbers (0-9) for known digits
- Use 'x' or 'X' as placeholders for unknown digits
- The number of placeholders is unlimited (but be aware that more placeholders will generate more combinations)

### Examples

1. Generate combinations for a number with three unknown digits:

```bash
clank -n 918115605xxx
```

2. Generate combinations for a number with two unknown digits:

```bash
clank 9181156xx99
```

## Project Structure

```
clank/
├── main.go        # Main source code
├── go.mod         # Go module file
└── README.md      # Documentation
```

## Technical Details

- Written in Go
- Dependencies:
  - Standard Go libraries:
    - `flag` for command-line argument parsing
    - `fmt` for I/O
    - `strings` for string manipulation
    - `os` for system operations
    - `net/http` for Truecaller API requests
    - `time` for rate limiting
- Truecaller API integration for number lookup

## Limitations

- Large numbers of placeholders (e.g., more than 6-7) may generate a very large number of combinations
- Processing time increases exponentially with the number of placeholders
- Memory usage scales with the number of combinations generated
- Truecaller API rate limits may apply
- Truecaller lookup requires internet connectivity

## Future Enhancements

Planned features for future releases:

- Progress bar for large combinations
- Output formatting options
- Country code validation
- Interactive mode
- Save results to file
- Integration with phone number lookup services
- Support for other search patterns

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the need for efficient phone number pattern matching
- Thanks to the Go community for the excellent standard library

## Author

atrey.dev

## Disclaimer

This tool is intended for legitimate use cases only. Please ensure you comply with all applicable laws and regulations regarding phone number lookup and privacy when using this tool.

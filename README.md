# Gadsby

Gadsby is a lightweight document search engine and local web server written in Go. It automatically scans for Microsoft Word (`.docx`) documents in a directory, indexes their full text and metadata, and serves an interactive search web interface.

## Features

- **Automated `.docx` Parsing**: Extracts text content and structured metadata (dates, reference numbers / *Aktenzeichen*, subject matter / *Rechtsgebiet*, decision status, keywords, and comments).
- **Fast Full-Text Search**: Powered by an in-memory search index ([Bluge](https://github.com/blugelabs/bluge)).
- **Tag Filtering**: Filter search results by extracted metadata tags.
- **Embedded Web UI**: Built with [Fiber](https://gofiber.io/) and HTML/JS static frontend embedded directly into the binary.
- **Document Downloads**: Easily preview metadata and download original `.docx` files directly from the search interface.

## Quick Start

### Prerequisites

- Go 1.18 or higher

### Installation & Execution

Clone the repository and run the application in any directory containing `.docx` files:

```bash
git clone https://github.com/jfrohnhofen/gadsby.git
cd gadsby
go build -o gadsby .
./gadsby
```

Once started, open your browser and navigate to:
```
http://localhost:9000
```

## How It Works

1. On startup, Gadsby recursively searches the working directory for all `.docx` files.
2. Each document is parsed to extract metadata fields and body text into a Bluge in-memory index.
3. The embedded web application listens on port `9000` and allows users to search document contents, apply tag filters, sort columns, and download files.

## License

This project is licensed under the [MIT License](LICENSE).

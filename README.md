# MyExtract

[![Go Reference](https://pkg.go.dev/badge/github.com/xavier268/myextract.svg)](https://pkg.go.dev/github.com/xavier268/myextract)
[![Go Report Card](https://goreportcard.com/badge/github.com/xavier268/myextract)](https://goreportcard.com/report/github.com/xavier268/myextract)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for extracting structured data from documents using Google's Gemini AI API. MyExtract supports multiple document formats and can return responses in JSON format using custom schemas or as plain text.

## Features

- **Multi-format Support**: Handles DOCX, TXT, MD, HTML, CSV, XML, RTF, PDF, and JSON files
- **Structured Output**: Extract data in JSON format using custom schemas
- **Automatic Text Extraction**: DOCX files are automatically converted to plain text
- **Flexible Configuration**: Customizable system prompts and output token limits
- **Resource Management**: Automatic cleanup of uploaded files to minimize API costs

## Installation

```bash
go get github.com/xavier268/myextract
```

## Prerequisites

- Go 1.18 or higher
- Google Gemini API key
- Dependencies:
  - `github.com/xavier268/mydocx` (for DOCX text extraction)
  - `google.golang.org/genai` (Google Generative AI client)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/xavier268/myextract"
    "google.golang.org/genai"
)

func main() {
    // Create extractor with your API key
    extractor, err := myextract.NewExtractor(context.Background(), "your-gemini-api-key")
    if err != nil {
        log.Fatal(err)
    }
    defer extractor.Close() // Important: always close to clean up resources
    
    // Upload a document
    err = extractor.Upload("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    // Extract data with a simple prompt
    result, err := extractor.Extract(nil, "Summarize the main points of this document")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result)
}
```

## Detailed Usage

### Creating an Extractor

```go
// Basic creation
extractor, err := myextract.NewExtractor(context.Background(), "your-api-key")

// With custom context
ctx := context.WithTimeout(context.Background(), 30*time.Second)
extractor, err := myextract.NewExtractor(ctx, "your-api-key")
```

### Configuration Options

```go
// Set maximum output tokens
extractor.SetMaxOutputToken(1000)

// Set system instructions
extractor.SetSystemPrompt("You are a helpful assistant that extracts key information from documents.")
```

### Uploading Files

The library supports various file formats:

```go
// Upload different file types
err := extractor.Upload("report.pdf")
err := extractor.Upload("data.csv")
err := extractor.Upload("document.docx")  // Automatically extracts text
err := extractor.Upload("webpage.html")
```

**Supported formats:**
- `.docx` - Microsoft Word (text extracted automatically)
- `.txt` - Plain text
- `.md` - Markdown
- `.html`, `.htm` - HTML
- `.csv` - Comma-separated values
- `.xml` - XML
- `.rtf` - Rich Text Format
- `.pdf` - PDF documents
- `.json` - JSON files

### Structured Data Extraction

Use schemas to get structured JSON responses:

```go
// Define a schema for structured output
schema := &genai.Schema{
    Type: genai.TypeObject,
    Properties: map[string]*genai.Schema{
        "title": {
            Type:        genai.TypeString,
            Description: "Document title",
        },
        "summary": {
            Type:        genai.TypeString,
            Description: "Brief summary of the document",
        },
        "key_points": {
            Type: genai.TypeArray,
            Items: &genai.Schema{
                Type: genai.TypeString,
            },
            Description: "Main points or takeaways",
        },
        "sentiment": {
            Type: genai.TypeString,
            Enum: []string{"positive", "negative", "neutral"},
            Description: "Overall sentiment of the document",
        },
    },
    Required: []string{"title", "summary"},
}

// Extract structured data
jsonResult, err := extractor.Extract(schema, "Extract the key information from this document")
```

### Plain Text Extraction

For simple text responses, pass `nil` as the schema:

```go
// Get plain text response
textResult, err := extractor.Extract(nil, "What are the main topics discussed in this document?")
```

## Complete Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/xavier268/myextract"
    "google.golang.org/genai"
)

func main() {
    // Initialize extractor
    extractor, err := myextract.NewExtractor(context.Background(), "your-gemini-api-key")
    if err != nil {
        log.Fatal("Failed to create extractor:", err)
    }
    defer extractor.Close()
    
    // Configure extractor
    extractor.
        SetMaxOutputToken(2000).
        SetSystemPrompt("You are an expert document analyzer. Extract information accurately and concisely.")
    
    // Upload multiple files
    files := []string{"contract.pdf", "report.docx", "data.csv"}
    for _, file := range files {
        if err := extractor.Upload(file); err != nil {
            log.Printf("Warning: Failed to upload %s: %v", file, err)
        }
    }
    
    // Define extraction schema
    schema := &genai.Schema{
        Type: genai.TypeObject,
        Properties: map[string]*genai.Schema{
            "document_type": {
                Type:        genai.TypeString,
                Description: "Type of document (contract, report, data, etc.)",
            },
            "key_entities": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Important names, dates, amounts, or other entities",
            },
            "action_items": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Any action items or next steps mentioned",
            },
        },
    }
    
    // Extract structured data
    result, err := extractor.Extract(schema, "Analyze these documents and extract key information")
    if err != nil {
        log.Fatal("Extraction failed:", err)
    }
    
    // Parse and display results
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(result), &data); err != nil {
        log.Fatal("Failed to parse JSON:", err)
    }
    
    fmt.Printf("Extracted data: %+v\n", data)
}
```

## Error Handling

The library returns standard Go errors. Common error scenarios:

- **Unsupported file types**: The library will return an error for unsupported formats
- **Upload failures**: Network issues or API limits may cause upload failures
- **Extraction failures**: Invalid prompts or API issues may cause extraction to fail

```go
if err := extractor.Upload("unsupported.xyz"); err != nil {
    fmt.Printf("Upload error: %v\n", err)
    // Handle error appropriately
}
```

## Resource Management

**Important**: Always call `Close()` to clean up uploaded files and avoid unnecessary API costs:

```go
extractor, err := myextract.NewExtractor(ctx, apiKey)
if err != nil {
    return err
}
defer extractor.Close() // This deletes uploaded files from Gemini API
```

The `Close()` method:
- Deletes all uploaded files from the Gemini API
- Prevents accumulating storage costs
- Is idempotent (safe to call multiple times)

## API Reference

### Types

#### `Extractor`
Main struct for document extraction operations.

### Functions

#### `NewExtractor(ctx context.Context, APIKey string) (*Extractor, error)`
Creates a new extractor instance with the provided API key.

#### `(*Extractor) SetMaxOutputToken(nb int) *Extractor`
Sets the maximum number of output tokens. Returns the extractor for method chaining.

#### `(*Extractor) SetSystemPrompt(systInstr string) *Extractor`
Sets system instructions for the AI model. Returns the extractor for method chaining.

#### `(*Extractor) Upload(filePath string) error`
Uploads a file to be used in extraction. Supports multiple formats.

#### `(*Extractor) Extract(schema *genai.Schema, prompt string) (string, error)`
Extracts information from uploaded files using the provided prompt and optional schema.

#### `(*Extractor) Close() error`
Cleans up resources and deletes uploaded files from the API.

## Best Practices

1. **Always close the extractor** to avoid accumulating API storage costs
2. **Use schemas** for structured data extraction when you need consistent output format
3. **Handle errors appropriately** - file uploads and API calls can fail
4. **Set reasonable token limits** to control API costs
5. **Use system prompts** to improve extraction quality and consistency

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Dependencies

- [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) - Google Generative AI Go client
- [github.com/xavier268/mydocx](https://github.com/xavier268/mydocx) - DOCX text extraction library
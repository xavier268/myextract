# MyExtract

[![Go Reference](https://pkg.go.dev/badge/github.com/xavier268/myextract.svg)](https://pkg.go.dev/github.com/xavier268/myextract)
[![Go Report Card](https://goreportcard.com/badge/github.com/xavier268/myextract)](https://goreportcard.com/report/github.com/xavier268/myextract)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

MyExtract is a powerful Go library that leverages Google's Gemini AI API to extract structured data from various document formats. The library intelligently processes documents and returns information in either structured JSON format using custom schemas or as plain text responses.

## Key Features

- **Universal Document Support**: Process DOCX, TXT, MD, HTML, CSV, XML, RTF, PDF, and JSON files seamlessly
- **Smart Structured Output**: Extract data in JSON format with user-defined schemas for consistent results
- **Intelligent Text Processing**: Automatically converts DOCX files to plain text for optimal processing
- **Configurable Settings**: Customize system instructions and output token limits to meet your needs
- **Efficient Resource Management**: Automatic cleanup of uploaded files to minimize API usage costs

## Installation

Install MyExtract using Go modules:

```bash
go get github.com/xavier268/myextract
```

## Requirements

- Go version 1.18 or higher
- Valid Google Gemini API key
- Required dependencies (automatically installed):
  - `github.com/xavier268/mydocx` for DOCX text extraction
  - `google.golang.org/genai` for Google Generative AI client integration

## Getting Started

Here's a simple example to get you started with MyExtract:

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
    // Initialize the extractor with your Gemini API key
    extractor, err := myextract.NewExtractor(context.Background(), "your-gemini-api-key")
    if err != nil {
        log.Fatal("Failed to create extractor:", err)
    }
    defer extractor.Close() // Always close to clean up resources
    
    // Upload a document for processing
    err = extractor.Upload("sample-document.pdf")
    if err != nil {
        log.Fatal("Failed to upload document:", err)
    }
    
    // Extract information using a natural language prompt
    result, err := extractor.Extract(nil, "Please summarize the key points from this document")
    if err != nil {
        log.Fatal("Failed to extract data:", err)
    }
    
    fmt.Println("Extracted summary:", result)
}
```

## Advanced Usage

### Initializing the Extractor

Create an extractor instance with different context configurations:

```go
// Basic initialization
extractor, err := myextract.NewExtractor(context.Background(), "your-api-key")

// With timeout context for better control
ctx := context.WithTimeout(context.Background(), 30*time.Second)
extractor, err := myextract.NewExtractor(ctx, "your-api-key")
```

### Customizing Extractor Behavior

Configure the extractor to match your specific requirements:

```go
// Set the maximum number of tokens in the response
extractor.SetMaxOutputToken(1500)

// Define custom system instructions to guide the AI's behavior
extractor.SetSystemPrompt("You are a professional document analyst. Extract information with high accuracy and provide clear, concise responses.")
```

### Uploading Documents

MyExtract supports a wide range of document formats:

```go
// Upload various document types
err := extractor.Upload("financial-report.pdf")
err := extractor.Upload("customer-data.csv")
err := extractor.Upload("meeting-notes.docx")  // Text is automatically extracted
err := extractor.Upload("web-content.html")
```

**Supported File Formats:**
- `.docx` - Microsoft Word documents (automatically converted to plain text)
- `.txt` - Plain text files
- `.md` - Markdown documents
- `.html`, `.htm` - HTML web pages
- `.csv` - Comma-separated value files
- `.xml` - XML documents
- `.rtf` - Rich Text Format files
- `.pdf` - Portable Document Format files
- `.json` - JSON data files

### Extracting Structured Data

Define schemas to receive structured JSON responses that match your data requirements:

```go
// Create a schema for consistent, structured output
schema := &genai.Schema{
    Type: genai.TypeObject,
    Properties: map[string]*genai.Schema{
        "document_title": {
            Type:        genai.TypeString,
            Description: "The main title or heading of the document",
        },
        "executive_summary": {
            Type:        genai.TypeString,
            Description: "A concise summary of the document's main content",
        },
        "key_findings": {
            Type: genai.TypeArray,
            Items: &genai.Schema{
                Type: genai.TypeString,
            },
            Description: "List of important discoveries, conclusions, or insights",
        },
        "overall_sentiment": {
            Type: genai.TypeString,
            Enum: []string{"positive", "negative", "neutral", "mixed"},
            Description: "The general tone or sentiment expressed in the document",
        },
        "priority_level": {
            Type: genai.TypeString,
            Enum: []string{"high", "medium", "low"},
            Description: "Urgency or importance level of the document content",
        },
    },
    Required: []string{"document_title", "executive_summary"},
}

// Extract data using the defined schema
structuredResult, err := extractor.Extract(schema, "Analyze this document and extract the key information according to the provided structure")
```

### Simple Text Extraction

For straightforward text responses without structured formatting:

```go
// Get a plain text response by passing nil as the schema
plainTextResult, err := extractor.Extract(nil, "What are the three most important topics discussed in this document?")
```

## Comprehensive Example

This example demonstrates a complete workflow with multiple files and advanced configuration:

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
    // Create extractor instance
    extractor, err := myextract.NewExtractor(context.Background(), "your-gemini-api-key")
    if err != nil {
        log.Fatal("Failed to initialize extractor:", err)
    }
    defer extractor.Close()
    
    // Configure extractor settings
    extractor.
        SetMaxOutputToken(3000).
        SetSystemPrompt("You are an expert business analyst. Analyze documents thoroughly and provide detailed, actionable insights.")
    
    // Upload multiple documents for comprehensive analysis
    documentPaths := []string{
        "quarterly-report.pdf",
        "market-analysis.docx",
        "customer-feedback.csv",
        "competitor-research.html",
    }
    
    for _, path := range documentPaths {
        if err := extractor.Upload(path); err != nil {
            log.Printf("Warning: Could not upload %s - %v", path, err)
        } else {
            fmt.Printf("Successfully uploaded: %s\n", path)
        }
    }
    
    // Define comprehensive extraction schema
    analysisSchema := &genai.Schema{
        Type: genai.TypeObject,
        Properties: map[string]*genai.Schema{
            "document_types": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Categories of documents analyzed (report, analysis, feedback, etc.)",
            },
            "business_insights": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Key business insights and strategic observations",
            },
            "action_recommendations": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Specific actions or next steps recommended based on the analysis",
            },
            "risk_factors": {
                Type: genai.TypeArray,
                Items: &genai.Schema{
                    Type: genai.TypeString,
                },
                Description: "Potential risks or concerns identified in the documents",
            },
            "confidence_score": {
                Type:        genai.TypeNumber,
                Description: "Confidence level in the analysis (0.0 to 1.0)",
            },
        },
        Required: []string{"business_insights", "action_recommendations"},
    }
    
    // Perform comprehensive extraction
    analysisResult, err := extractor.Extract(
        analysisSchema, 
        "Perform a comprehensive business analysis of all uploaded documents. Identify key insights, risks, and provide actionable recommendations.",
    )
    if err != nil {
        log.Fatal("Analysis failed:", err)
    }
    
    // Parse and display structured results
    var businessAnalysis map[string]interface{}
    if err := json.Unmarshal([]byte(analysisResult), &businessAnalysis); err != nil {
        log.Fatal("Failed to parse analysis results:", err)
    }
    
    // Display results in a user-friendly format
    fmt.Println("\n=== Business Analysis Results ===")
    prettyJSON, _ := json.MarshalIndent(businessAnalysis, "", "  ")
    fmt.Println(string(prettyJSON))
}
```

## Error Handling Best Practices

Implement robust error handling for reliable operation:

```go
// Handle file upload errors gracefully
if err := extractor.Upload("document.txt"); err != nil {
    switch {
    case strings.Contains(err.Error(), "unsupported"):
        fmt.Println("File format not supported - please use PDF, DOCX, or TXT")
    case strings.Contains(err.Error(), "not found"):
        fmt.Println("File not found - please check the file path")
    default:
        fmt.Printf("Upload failed: %v\n", err)
    }
    return
}

// Handle extraction errors with specific responses
result, err := extractor.Extract(schema, prompt)
if err != nil {
    if strings.Contains(err.Error(), "token limit") {
        fmt.Println("Response too long - try reducing MaxOutputToken or simplifying the prompt")
    } else if strings.Contains(err.Error(), "quota") {
        fmt.Println("API quota exceeded - please check your Gemini API usage")
    } else {
        fmt.Printf("Extraction error: %v\n", err)
    }
    return
}
```

## Resource Management

Proper resource management is crucial for cost control and performance:

```go
extractor, err := myextract.NewExtractor(ctx, apiKey)
if err != nil {
    return fmt.Errorf("extractor creation failed: %w", err)
}
defer extractor.Close() // Critical: always close to prevent resource leaks

// The Close() method performs these important cleanup tasks:
// - Removes all uploaded files from Gemini API storage
// - Prevents accumulation of storage charges
// - Can be called multiple times safely (idempotent operation)
```

## API Documentation

### Core Types

#### `Extractor`
The main struct that handles all document extraction operations and manages API interactions.

### Primary Functions

#### `NewExtractor(ctx context.Context, APIKey string) (*Extractor, error)`
Creates and initializes a new extractor instance using the provided Gemini API key.

**Parameters:**
- `ctx`: Context for controlling request lifecycle and timeouts
- `APIKey`: Valid Google Gemini API key for authentication

**Returns:** Configured extractor instance or error if initialization fails

#### `(*Extractor) SetMaxOutputToken(tokenCount int) *Extractor`
Configures the maximum number of tokens allowed in AI responses. Useful for controlling response length and API costs.

**Parameters:**
- `tokenCount`: Maximum number of output tokens (typically 100-4000)

**Returns:** The extractor instance for method chaining

#### `(*Extractor) SetSystemPrompt(instructions string) *Extractor`
Sets system-level instructions that guide the AI model's behavior and response style.

**Parameters:**
- `instructions`: Detailed instructions for how the AI should analyze and respond

**Returns:** The extractor instance for method chaining

#### `(*Extractor) Upload(filePath string) error`
Uploads a document file to the Gemini API for processing. Supports multiple file formats.

**Parameters:**
- `filePath`: Path to the document file to upload

**Returns:** Error if upload fails due to unsupported format, network issues, or API limits

#### `(*Extractor) Extract(schema *genai.Schema, prompt string) (string, error)`
Extracts information from all uploaded files using the specified prompt and optional response schema.

**Parameters:**
- `schema`: Optional schema for structured JSON output (pass nil for plain text)
- `prompt`: Natural language instruction describing what information to extract

**Returns:** Extracted information as string (JSON if schema provided) or error

#### `(*Extractor) Close() error`
Cleans up all resources and removes uploaded files from the Gemini API to prevent storage costs.

**Returns:** Error if cleanup fails (safe to ignore in most cases)

## Best Practices and Recommendations

### Cost Optimization
1. **Always call Close()** - Prevents accumulating storage charges from uploaded files
2. **Set appropriate token limits** - Use SetMaxOutputToken() to control response length and costs
3. **Upload files efficiently** - Only upload files that are needed for the current extraction task
4. **Use structured schemas** - Define specific schemas to get only the data you need

### Performance Optimization
1. **Use context timeouts** - Set reasonable timeouts to prevent hanging requests
2. **Handle errors gracefully** - Implement retry logic for transient failures
3. **Batch related extractions** - Upload multiple related files once for comprehensive analysis
4. **Optimize prompts** - Write clear, specific prompts to get better results faster

### Quality Improvement
1. **Use system prompts** - Set clear instructions to improve consistency and accuracy
2. **Define detailed schemas** - Specify required fields and data types for structured output
3. **Validate responses** - Check extracted data against expected formats and ranges
4. **Test with sample data** - Verify extraction quality with known documents before production use

## Contributing to MyExtract

We welcome contributions from the community! Here's how you can help:

1. **Report Issues** - Submit bug reports with detailed reproduction steps
2. **Suggest Features** - Propose new functionality or improvements
3. **Submit Pull Requests** - Contribute code improvements or new features
4. **Improve Documentation** - Help make the documentation clearer and more comprehensive

Please review our contribution guidelines and code of conduct before submitting pull requests.

## License Information

This project is licensed under the MIT License, which allows for both personal and commercial use with minimal restrictions. See the LICENSE file for complete license terms.

## Dependencies and Credits

MyExtract relies on these excellent open-source libraries:

- **[google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai)** - Official Google Generative AI Go client library
- **[github.com/xavier268/mydocx](https://github.com/xavier268/mydocx)** - Specialized library for extracting text from Microsoft Word DOCX files

## Version History and Changelog

Check the releases page for detailed information about version updates, new features, and bug fixes.
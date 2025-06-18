package myextract

// See : https://ai.google.dev/gemini-api/docs/structured-output?hl=fr#generating-enums

import (
	"context"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xavier268/mydocx"
	"google.golang.org/genai"
)

// Extractor represents a document extraction client that interfaces with Google's Gemini API
// to extract structured or unstructured data from various document formats.
type Extractor struct {
	client    *genai.Client   // Gemini API client for making requests
	ctx       context.Context // Context for controlling request lifecycle
	model     string          // Gemini model name to use for extraction
	systInstr *genai.Content  // System instructions to guide the AI's behavior
	maxToken  int32           // Maximum number of tokens to generate in response
	files     []*genai.File   // Collection of uploaded files available for extraction
	log       log.Logger      // Logger for tracking operations and debugging
}

// NewExtractor creates a new Extractor instance with the provided API key and context.
// If ctx is nil, it defaults to context.Background().
// The extractor must be closed when finished to clean up uploaded files and avoid costs.
// Make sure you close created extractor when finished (used defer !)
func NewExtractor(ctx context.Context, APIKey string) (*Extractor, error) {
	// Use background context if none provided
	if ctx == nil {
		ctx = context.Background()
	}

	// Initialize Gemini API client with provided configuration
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	// Return configured extractor instance
	return &Extractor{
		client: c,
		ctx:    ctx,
		model:  MODEL, // MODEL constant should be defined elsewhere in the package
		log:    *log.Default(),
	}, nil
}

// SetMaxOutputToken sets the maximum number of tokens the AI can generate in its response.
// This helps control API costs and response length. Returns the extractor for method chaining.
// Negative values are converted to 0.
func (e *Extractor) SetMaxOutputToken(nb int) *Extractor {
	// Ensure non-negative value using max function
	e.maxToken = max(int32(nb), 0)
	return e
}

// SetSystemPrompt sets system-level instructions that guide the AI's behavior during extraction.
// These instructions are applied to all subsequent Extract calls.
// An empty string clears the system prompt. Returns the extractor for method chaining.
func (e *Extractor) SetSystemPrompt(systInstr string) *Extractor {
	if systInstr != "" {
		// Create system instruction content with user role
		e.systInstr = genai.NewContentFromText(systInstr, genai.RoleUser)
	} else {
		// Clear system instructions if empty string provided
		e.systInstr = nil
	}
	return e
}

// Extract performs data extraction from all uploaded files using the provided prompt.
// If a schema is provided, the response will be structured JSON conforming to that schema.
// If schema is nil, the response will be plain text.
// All files previously uploaded to this extractor are included in the extraction context.
func (e *Extractor) Extract(schema *genai.Schema, prompt string) (result string, err error) {
	var config *genai.GenerateContentConfig

	// Configure response format based on whether schema is provided
	// response in json if schema specified
	if schema != nil {
		// Structured JSON response with provided schema
		config = &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema}
	} else {
		// Plain text response when no schema specified
		// réponse texte si schema nil
		config = &genai.GenerateContentConfig{
			ResponseMIMEType: "text/plain",
		}
	}

	// Apply token limit if set
	if e.maxToken > 0 {
		config.MaxOutputTokens = e.maxToken
	}

	// Apply system instructions if set
	if e.systInstr != nil {
		config.SystemInstruction = e.systInstr
	}

	// Build prompt parts including all uploaded files and the text prompt
	// Select the uploaded files
	promptParts := make([]*genai.Part, 0, len(e.files)+1)

	// Add each uploaded file as a URI part
	for _, f := range e.files {
		promptParts = append(promptParts, genai.NewPartFromURI(f.URI, f.MIMEType))
	}

	// Add the text prompt as the final part
	// Add the prompt, create the content input
	promptParts = append(promptParts, genai.NewPartFromText(prompt))

	// Create content structure for the API call
	contents := []*genai.Content{
		genai.NewContentFromParts(promptParts, genai.RoleUser),
	}

	// Make the API call to generate content
	// query
	r, err := e.client.Models.GenerateContent(e.ctx, e.model, contents, config)
	if err != nil {
		return "", err
	}

	// Return the generated text response
	return r.Text(), nil
}

// Close cleans up all resources associated with the extractor.
// This includes deleting all uploaded files from the Gemini API to prevent ongoing storage costs.
// This method is idempotent and can be called multiple times safely.
// It's critical to call this method to avoid accumulating API storage charges.
func (e *Extractor) Close() error {
	var ee []string // collect errors !

	// Iterate through all uploaded files and delete them from the API
	for _, f := range e.files {
		if f == nil {
			continue
		}

		// Log the deletion for debugging purposes
		e.log.Printf("Deleting (%s)\n", f.Name)

		// Delete file using background context to ensure deletion completes
		// even if the extractor's context is cancelled
		// don't use existing context, to ensure deletion ...
		_, err := e.client.Files.Delete(context.Background(), f.Name, nil)
		if err != nil {
			// Collect errors but continue deleting other files
			ee = append(ee, err.Error())
		}
	}

	// Clear the files slice to prevent double-deletion
	e.files = nil // reset files

	// Return any errors encountered during deletion
	if len(ee) == 0 {
		return nil
	} else {
		ee = nil
		return fmt.Errorf("error while deleting files : %v", ee)
	}
}

// Upload adds a file from the local filesystem to the extractor for use in subsequent extractions.
// The file type is automatically detected by extension and appropriate MIME type is set.
// Special handling for DOCX files: text is extracted before upload.
// Supported formats: DOCX, TXT, MD, HTML, HTM, CSV, XML, RTF, PDF, JSON
func (e *Extractor) Upload(filePath string) error {
	// Convert relative path to absolute path for consistency
	// Convert to absolute path
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Initialize upload configuration
	// detect mime type
	ufc := &genai.UploadFileConfig{}

	// Extract file extension and convert to uppercase for comparison
	ext := strings.ToUpper(filepath.Ext(filePath))

	// Special handling for DOCX files - extract text content first
	// Handle docx by extracting text first
	if ext == ".DOCX" {
		// Set MIME type for extracted text
		ufc.MIMEType = "text/plain"

		// Extract text from DOCX using mydocx library
		data, err := mydocx.ExtractText(filePath)
		if err != nil {
			return err
		}

		// Join extracted text lines into single content string
		cont := strings.Join(data["word/document.xml"], "\n")

		// Upload the extracted text content
		f, err := e.client.Files.Upload(e.ctx, strings.NewReader(cont), ufc)
		if err != nil {
			return err
		}

		// Add uploaded file to the collection and log success
		e.files = append(e.files, f)
		e.log.Printf("Uploaded (%s) : %q\n", f.Name, filePath)
		return nil
	}

	// For non-DOCX files, open the file for direct upload
	// now, we need to open file
	of, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer of.Close()

	// Set appropriate MIME type based on file extension
	// set mime type for recognized file types
	switch ext {
	case ".TXT":
		ufc.MIMEType = "text/plain"
	case ".MD":
		ufc.MIMEType = "text/md"
	case ".HTML", ".HTM":
		ufc.MIMEType = "text/html"
	case ".CSV":
		ufc.MIMEType = "text/csv"
	case ".XML":
		ufc.MIMEType = "text/xml"
	case ".RTF":
		ufc.MIMEType = "text/rtf"
	case ".PDF":
		ufc.MIMEType = "application/pdf"
	case ".JSON", ".JASON": // Note: .JASON handles common misspelling
		ufc.MIMEType = "application/json"
	default:
		// Return error for unsupported file types
		return fmt.Errorf("file type not supported : %v", filePath)
	}

	// Upload the file directly to the API
	// actual upload for non word files
	f, err := e.client.Files.Upload(e.ctx, of, ufc)
	if err != nil {
		return err
	}

	// Add uploaded file to the collection and log success
	e.files = append(e.files, f)
	e.log.Printf("Uploaded (%s) : %q\n", f.Name, filePath)
	return nil
}

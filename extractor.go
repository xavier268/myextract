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

type Extractor struct {
	client    *genai.Client
	ctx       context.Context
	model     string
	systInstr *genai.Content // instructions system
	maxToken  int32          // max output token
	files     []*genai.File  // files uploaded and used in queries
	log       log.Logger
}

// Make sure you cloase created extractor when not needed anymore !
func NewExtractor(ctx context.Context, APIKey string) (*Extractor, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &Extractor{
		client: c,
		ctx:    ctx,
		model:  MODEL,
		log:    *log.Default(),
	}, nil
}

// (re)Set Max Output tokens
func (e *Extractor) SetMaxOutputToken(nb int) *Extractor {
	e.maxToken = max(int32(nb), 0)
	return e
}

// (re)Set System prompt instructions
func (e *Extractor) SetSystemPrompt(systInstr string) *Extractor {
	if systInstr != "" {
		e.systInstr = genai.NewContentFromText(systInstr, genai.RoleUser)
	} else {
		e.systInstr = nil
	}
	return e
}

// Extrait les données en json, en utilisant le schema fourni.
// Si pas de schema, prompt standard, réponse en text.
// Utilise tous les fichiers uploadés dans l'extractor
func (e *Extractor) Extract(schema *genai.Schema, prompt string) (result string, err error) {
	var config *genai.GenerateContentConfig
	// response in json if schema specified
	if schema != nil {
		config = &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema}
	} else {
		// réponse texte si schema nil
		config = &genai.GenerateContentConfig{
			ResponseMIMEType: "text/plain",
		}
	}
	if e.maxToken > 0 {
		config.MaxOutputTokens = e.maxToken
	}
	if e.systInstr != nil {
		config.SystemInstruction = e.systInstr
	}

	// Select the uploaded files
	promptParts := make([]*genai.Part, 0, len(e.files)+1)
	for _, f := range e.files {
		promptParts = append(promptParts, genai.NewPartFromURI(f.URI, f.MIMEType))
	}
	// Add the prompt, create the content input
	promptParts = append(promptParts, genai.NewPartFromText(prompt))
	contents := []*genai.Content{
		genai.NewContentFromParts(promptParts, genai.RoleUser),
	}

	// query
	r, err := e.client.Models.GenerateContent(e.ctx, e.model, contents, config)
	if err != nil {
		return "", err
	}
	return r.Text(), nil
}

// Free all resources associated witg extractor.
// Required to save costs for uploaded files !
// Idempotent.
func (e *Extractor) Close() error {
	var ee []string // collect errors !
	for _, f := range e.files {
		if f == nil {
			continue
		}
		e.log.Printf("Deleting (%s)\n", f.Name)
		_, err := e.client.Files.Delete(context.Background(), f.Name, nil) // don't use existing context, to ensure deletion ...
		if err != nil {
			ee = append(ee, err.Error())
		}
	}
	e.files = nil // reset files
	if len(ee) == 0 {
		return nil
	} else {
		ee = nil
		return fmt.Errorf("error while deleting files : %v", ee)
	}
}

// Upload a file from local path on computer.
// *.docx files have their text extracted first.
// *.txt files are sent as is.
// *.PDF files are identified and transferred as is.
// Other files generate an error (for the moment ...)
func (e *Extractor) Upload(filePath string) error {
	// Convert to absolute path
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// detect mime type
	ufc := &genai.UploadFileConfig{}
	ext := strings.ToUpper(filepath.Ext(filePath))

	// Handle docx by extracting text first
	if ext == ".DOCX" {
		ufc.MIMEType = "text/plain"
		data, err := mydocx.ExtractText(filePath)
		if err != nil {
			return err
		}
		cont := strings.Join(data["word/document.xml"], "\n")
		f, err := e.client.Files.Upload(e.ctx, strings.NewReader(cont), ufc)
		if err != nil {
			return err
		}
		e.files = append(e.files, f)
		e.log.Printf("Uploaded (%s) : %q\n", f.Name, filePath)
		return nil
	}

	// now, we need to open file
	of, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer of.Close()

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
	case ".JSON", ".JASON":
		ufc.MIMEType = "application/json"
	default:
		return fmt.Errorf("file type not supported : %v", filePath)
	}

	// actual upload for non word files
	f, err := e.client.Files.Upload(e.ctx, of, ufc)
	if err != nil {
		return err
	}
	e.files = append(e.files, f)
	e.log.Printf("Uploaded (%s) : %q\n", f.Name, filePath)
	return nil
}

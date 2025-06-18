package myextract

import (
	"context"
	"path/filepath"
	"testing"

	"google.golang.org/genai"
)

// You google API key here
// Required only for testing
var TEST_KEY string

func TestKey(t *testing.T) {
	if len(TEST_KEY) == 0 {
		t.Fatal("You need to set the TEST_KEY variable")
	}
}

func TestExtractorPlainWithSystemPrompt(t *testing.T) {
	e, err := NewExtractor(context.Background(), TEST_KEY)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	e.SetSystemPrompt("You only speak German, even if questions are raised in another langage !")
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.Extract(nil, "Bonjour, explique moi ce que tu sais faire ?")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestExtractorJson(t *testing.T) {
	e, err := NewExtractor(context.Background(), TEST_KEY)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	schema := &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"recipeName": {Type: genai.TypeString},
				"ingredients": {
					Type:  genai.TypeArray,
					Items: &genai.Schema{Type: genai.TypeString},
				},
			},
			PropertyOrdering: []string{"recipeName", "ingredients"},
		}}

	r, err := e.Extract(schema, "Donne moi 3 bonnes recettes de cuisine pour les grosses chaleurs.")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
func TestExtractorJsonWithTruncatedOutput(t *testing.T) {
	e, err := NewExtractor(context.Background(), TEST_KEY)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	schema := &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"recipeName": {Type: genai.TypeString},
				"ingredients": {
					Type:  genai.TypeArray,
					Items: &genai.Schema{Type: genai.TypeString},
				},
			},
			PropertyOrdering: []string{"recipeName", "ingredients"},
		}}
	e.SetMaxOutputToken(20)
	r, err := e.Extract(schema, "Donne moi 3 bonnes recettes de cuisine pour les grosses chaleurs.")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestUpload(t *testing.T) {
	e, err := NewExtractor(context.Background(), TEST_KEY)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	err = e.Upload(filepath.Join("testFiles", "pdf.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	err = e.Upload(filepath.Join("testFiles", "word.docx"))
	if err != nil {
		t.Fatal(err)
	}
	err = e.Upload(filepath.Join("testFiles", "html.html"))
	if err != nil {
		t.Fatal(err)
	}
	err = e.Upload(filepath.Join("testFiles", "csv.csv"))
	if err != nil {
		t.Fatal(err)
	}
	err = e.Upload(filepath.Join("testFiles", "txt.txt"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.Extract(nil, "Que contiennent ces fichiers ? Résume les en 3 lignes par fichier.")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

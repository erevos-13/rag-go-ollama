package models

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
)

type Document struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func UpdateDocument(c *gin.Context) error {
	ctx := context.Background()

	documentID := c.PostForm("id")
	if documentID == "" {
		return fmt.Errorf("document ID is required")
	}

	documentTitle, exists := c.GetPostForm("title")
	if !exists {
		documentTitle = "Untitled Document"
	}

	documentVersion := 1
	if versionStr, exists := c.GetPostForm("version"); exists {
		if v, err := fmt.Sscanf(versionStr, "%d", &documentVersion); err != nil || v == 0 {
			documentVersion = 1
		}
	}

	fmt.Printf("Processing document: ID=%s, Title=%s, Version=%d\n",
		documentID, documentTitle, documentVersion)

	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println("Error getting form file: ", err)
		return err
	}

	dirName := fmt.Sprintf("doc-update-%d-%s", time.Now().Unix(), file.Filename)
	tempDir, err := os.MkdirTemp("", dirName)
	if err != nil {
		fmt.Println("Error creating temp directory: ", err)
		return err
	}
	defer os.RemoveAll(tempDir)

	uploadPath := filepath.Join(tempDir, file.Filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		fmt.Println("Error saving uploaded file: ", err)
		return err
	}

	uploadedFile, err := os.Open(uploadPath)
	if err != nil {
		fmt.Println("Error opening uploaded file: ", err)
		return err
	}
	defer uploadedFile.Close()

	fileInfo, err := uploadedFile.Stat()
	if err != nil {
		fmt.Println("Error getting file info: ", err)
		return err
	}

	connManager := GetConnectionManager()

	loader := documentloaders.NewPDF(uploadedFile, fileInfo.Size())
	docs, err := loader.Load(context.Background())
	if err != nil {
		return err
	}

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(100),
		textsplitter.WithSeparators([]string{". "}),
	)
	var texts []string
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
	splitTexts, err := splitter.SplitText(strings.Join(texts, " "))
	if err != nil {
		return err
	}

	// Convert strings to schema.Document objects
	docs_to_store := make([]schema.Document, len(splitTexts))
	for i, text := range splitTexts {
		docs_to_store[i] = schema.Document{
			PageContent: text,
			Metadata: map[string]interface{}{
				"source":         uploadPath,
				"document_id":    documentID,
				"document_title": documentTitle,
			},
		}
	}

	store, err := connManager.GetChromaClient()
	if err != nil {
		fmt.Println("Error getting Chroma client: ", err)
		return err
	}

	_, err = store.AddDocuments(ctx, docs_to_store, vectorstores.WithNameSpace("doc_search"))
	if err != nil {
		return err
	}

	return nil
}

func SearchDocument(query string, c *gin.Context) (string, []schema.Document, error) {
	ctx := c.Request.Context()

	connManager := GetConnectionManager()

	store, err := connManager.GetChromaClient()
	if err != nil {
		return "", nil, fmt.Errorf("failed to connect to Chroma: %v", err)
	}

	llm, err := connManager.GetOllamaLLMClient()
	if err != nil {
		return "", nil, err
	}

	const maxRetrievalCount = 2
	results, err := store.SimilaritySearch(ctx, query, maxRetrievalCount, vectorstores.WithNameSpace("doc_search"))
	if err != nil {
		return "", nil, err
	}

	const maxContextDocs = 2
	docsToInclude := results
	if len(results) > maxContextDocs {
		docsToInclude = results[:maxContextDocs]
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Based on the following information:\n\n")
	for i, doc := range docsToInclude {
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.PageContent))
	}

	fullPrompt := fmt.Sprintf(`%s
	Question: %s
	Please provide a comprehensive answer based on the context provided.
	Please answer in a way that is easy to understand and helpful to the user.
	`,
		contextBuilder.String(), query)

	const temperature = 0.5
	completion, err := llm.Call(ctx, fullPrompt,
		llms.WithTemperature(temperature),
	)
	if err != nil {
		return "", nil, err
	}
	return completion, results, nil
}

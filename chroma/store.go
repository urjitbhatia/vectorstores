package chroma

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	gochroma "github.com/urjitbhatia/gochroma"
	"math"
)

type Store struct {
	client     gochroma.Chroma
	collection gochroma.Collection
	chromaOptions
}

func NewChroma(opts ...Options) (*Store, error) {
	chromaOpts, err := applyOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot init Chroma Client: %w", err)
	}
	client, err := gochroma.NewClient(chromaOpts.addr)
	if err != nil {
		return nil, err
	}

	col, err := client.GetOrCreateCollection(
		chromaOpts.collectionName,
		chromaOpts.distanceFn,
		chromaOpts.collectionMetadata)
	if err != nil {
		return nil, err
	}

	return &Store{
		client:        client,
		collection:    col,
		chromaOptions: chromaOpts,
	}, nil
}

// AddDocuments to the Chroma collection
func (c *Store) AddDocuments(_ context.Context, documents []schema.Document, option ...vectorstores.Option) error {
	opts := applyOperationOpts()
	// custom embedder if supplied
	embedder := c.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	chromaDocs := make([]gochroma.Document, len(documents))
	for i, doc := range documents {
		chromaDocs[i].Metadata = doc.Metadata
		chromaDocs[i].Content = doc.PageContent
		chromaDocs[i].ID = uuid.NewString()
	}
	if err := c.collection.Add(chromaDocs, embedder); err != nil {
		return fmt.Errorf("unable to add documents to the store. Err: %w", err)
	}
	return nil
}

// GetAllDocuments returns *all* the documents in the collection. Useful for debugging, but be careful
// while using in production.
func (c *Store) GetAllDocuments() ([]schema.Document, error) {
	docs, err := c.collection.Get(nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed fetching documents from store: %w", err)
	}
	schemaDocs := make([]schema.Document, len(docs))
	for i, doc := range docs {
		schemaDocs[i] = schema.Document{
			PageContent: doc.Content,
			Metadata:    doc.Metadata,
		}
	}
	return schemaDocs, nil
}

// SimilaritySearch Performs a similarity search on the given query within the collection targeted
// by this client.
// Note: If a distance Threshold is set, the number of documents returned can be *lower* than the
// requested numDocuments
func (c *Store) SimilaritySearch(_ context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	opts := applyOperationOpts()
	// custom embedder if supplied
	embedder := c.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	results, err := c.collection.Query(query,
		int32(numDocuments), nil, nil, nil, embedder)
	if err != nil {
		return nil, err
	}

	maxDistance := float32(math.Max(math.MaxFloat32, float64(opts.ScoreThreshold)))

	var docs []schema.Document
	for _, r := range results {
		if r.Distance > maxDistance {
			continue
		}
		docs = append(docs, schema.Document{
			PageContent: r.Content,
			Metadata:    r.Metadata,
			Score:       r.Distance,
		})
	}
	return docs, nil
}

func (c *Store) CollectionSize() (int, error) {
	return c.collection.Count()
}

// CollectionName Returns the collection name associated with this instance of the Chroma client
func (c *Store) CollectionName() string {
	return c.collectionName
}

// ResetDB Resets the Chroma DB.
// See: https://docs.trychroma.com/reference/Client#reset
func (c *Store) ResetDB() error {
	_, err := c.client.Reset()
	return err
}

// ClearCollection Drops and recreates the collection. All documents will be lost
// The original DistanceFn and Metadata are maintained
func (c *Store) ClearCollection() error {
	err := c.client.DeleteCollection(c.collectionName)
	if err != nil {
		return err
	}

	col, err := c.client.CreateCollection(c.collectionName, c.distanceFn, c.collectionMetadata)
	if err != nil {
		return err
	}
	c.collection = col
	return nil
}

func applyOperationOpts(options ...vectorstores.Option) *vectorstores.Options {
	opts := &vectorstores.Options{}
	for _, o := range options {
		o(opts)
	}
	return opts
}

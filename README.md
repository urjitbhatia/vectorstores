# vectorstores

[![Go Reference](https://pkg.go.dev/badge/github.com/urjitbhatia/vectorstores.svg)](https://pkg.go.dev/github.com/urjitbhatia/vectorstores)

This repo is a collection of clients for various vector stores aligned with the `VectorStores` interface
defined in https://github.com/tmc/langchaingo. This repo allows for testing the store client implementations.

## Supported Stores

1. Chroma: https://docs.trychroma.com/
2. PGVector: https://github.com/pgvector/pgvector
   - Make sure the Postgres version you are using has the [PGVector extension supported](https://github.com/pgvector/pgvector#installation)

## Usage:

```go
// Initialize the PGVector store
store, err = pgvector.NewPGVectorStore(
   pgvector.WithCreds("test", "test"),
   pgvector.WithCollection("myproject"),
   pgvector.WithEndpoint("localhost", 5432),
   pgvector.WithDBName("test"),
   pgvector.WithEmbedder(openAIEmbedder),
)
// Or the Chroma Store
store, err = chroma.NewChroma(
   chroma.WithAddr("http://localhost:8000"),
   chroma.WithEmbedder(testEmbedder{}),
   chroma.WithCollection("unittest", nil),
)

// Add documents
err := store.AddDocuments(context.Background(), docs)

// Search for documents
docs, err := store.SimilaritySearch(context.Background(), query, 1)
```
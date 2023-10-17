package store

import "github.com/tmc/langchaingo/vectorstores"

type VectorStore interface {
	vectorstores.VectorStore
	CollectionSize() (int, error)
	CollectionName() string
	ClearCollection() error
}

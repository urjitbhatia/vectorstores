package pgvector

import (
	"database/sql"
	"github.com/tmc/langchaingo/embeddings"
)

type pgvOptions struct {
	DB             *sql.DB
	Username       string
	Password       string
	DBName         string
	SSLMode        string
	EmbeddingModel EmbeddingModel
	Embedder       embeddings.Embedder
	TableName      string
	Host           string
	Port           int
}
type Options func(o *pgvOptions)

func WithClient(sqlClient *sql.DB) Options {
	return func(o *pgvOptions) {
		o.DB = sqlClient
	}
}

func WithCreds(user, pwd string) Options {
	return func(o *pgvOptions) {
		o.Username = user
		o.Password = pwd
	}
}

func WithEndpoint(host string, port int) Options {
	return func(o *pgvOptions) {
		o.Host = host
		o.Port = port
	}
}

func WithEmbedder(e embeddings.Embedder) Options {
	return func(o *pgvOptions) {
		o.Embedder = e
	}
}

func WithDBName(dbname string) Options {
	return func(o *pgvOptions) {
		o.DBName = dbname
	}
}

func WithCollection(name string) Options {
	return func(o *pgvOptions) {
		o.TableName = name
	}
}

func WithSSLMode(mode string) Options {
	return func(o *pgvOptions) {
		o.SSLMode = mode
	}
}

type EmbeddingModel int

const (
	EmbeddingOpenAI_ada_002 EmbeddingModel = iota
	EmbeddingOpenAI                        // default for openai
)

var VectorDimensionsByEmbeddingModel = map[EmbeddingModel]int{
	EmbeddingOpenAI:         1536,
	EmbeddingOpenAI_ada_002: 1536,
}

func WithEmbeddingMode(e EmbeddingModel) Options {
	return func(o *pgvOptions) {
		o.EmbeddingModel = e
	}
}

func applyOptions(opts ...Options) pgvOptions {
	o := pgvOptions{
		Host:    "localhost",
		Port:    5432,
		SSLMode: "disable",
	}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

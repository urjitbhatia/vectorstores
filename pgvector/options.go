package pgvector

import (
	"database/sql"
	"github.com/tmc/langchaingo/embeddings"
)

type pgvOptions struct {
	DB              *sql.DB
	Username        string
	Password        string
	DBName          string
	SSLMode         string
	EmbeddingModel  EmbeddingModel
	Embedder        embeddings.Embedder
	TableName       string
	Host            string
	Port            int
	CreateExtension bool
}
type Options func(o *pgvOptions)

// WithClient sets an existing sql client - useful for proxying or testing or supplying a pre-made
// sql client connection with customizations
func WithClient(sqlClient *sql.DB) Options {
	return func(o *pgvOptions) {
		o.DB = sqlClient
	}
}

// WithCreateExtension If set, will attempt to enable the PGVECTOR extension in the database.
// The user should have the appropriate permissions. Otherwise, provision the database with pgvector
// pre-enabled
func WithCreateExtension() Options {
	return func(o *pgvOptions) {
		o.CreateExtension = true
	}
}

// WithCreds sets the user creds for postgres
func WithCreds(user, pwd string) Options {
	return func(o *pgvOptions) {
		o.Username = user
		o.Password = pwd
	}
}

// WithEndpoint sets the host and port for the postgres instance to connect with
func WithEndpoint(host string, port int) Options {
	return func(o *pgvOptions) {
		o.Host = host
		o.Port = port
	}
}

// WithEmbedder sets the embeddings provider
func WithEmbedder(e embeddings.Embedder) Options {
	return func(o *pgvOptions) {
		o.Embedder = e
	}
}

// WithDBName sets the db name to use with postgres
func WithDBName(dbname string) Options {
	return func(o *pgvOptions) {
		o.DBName = dbname
	}
}

// WithCollection sets the name of the collection (table within the postgres database)
func WithCollection(name string) Options {
	return func(o *pgvOptions) {
		o.TableName = name
	}
}

// WithSSLMode only "require" (default), "verify-full", "verify-ca", and "disable" supported
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

// WithEmbeddingMode sets the embeddings mode. Currently, only OpenAI ADA_002 is supported.
func WithEmbeddingMode(e EmbeddingModel) Options {
	return func(o *pgvOptions) {
		o.EmbeddingModel = e
	}
}

func applyOptions(opts ...Options) pgvOptions {
	o := pgvOptions{
		Host:    "localhost",
		Port:    5432,
		SSLMode: "require",
	}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

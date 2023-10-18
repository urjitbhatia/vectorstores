package pgvector

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	store "github.com/urjitbhatia/vectorstores"
)

// Store is a postgres based implementation of a vector store
type Store struct {
	pgvOptions
}

func NewPGVectorStore(opts ...Options) (store.VectorStore, error) {
	o := applyOptions(opts...)
	// checks
	if o.TableName == "" {
		return nil, fmt.Errorf("collection name needs to be set")
	}

	if o.DB == nil {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			o.Host, o.Port,
			o.Username, o.Password,
			o.DBName, o.SSLMode)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to pg store. Err: %w", err)
		}
		o.DB = db
	}

	// verify or enable vector extensions
	if o.CreateExtension {
		if _, err := o.DB.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
			return nil, err
		}
	}
	// verify that vector extension is enabled
	ext := o.DB.QueryRow("SELECT COUNT(*) FROM pg_extension where extname = 'vector'")
	var count int
	err := ext.Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("vector extension is not enabled, cannot continue")
	}

	// create collection
	vectorDims := VectorDimensionsByEmbeddingModel[o.EmbeddingModel]
	collCreateQuery := fmt.Sprintf(
		`
		CREATE TABLE IF NOT EXISTS %s 
		(id serial PRIMARY KEY, embedding vector(%d), metadata JSON, content text)
		`,
		o.TableName, vectorDims)
	if _, err := o.DB.Exec(collCreateQuery); err != nil {
		return nil, err
	}

	// create index
	// Same as chroma
	// see: https://github.com/chroma-core/chroma/blob/0d675094035cf87904c77404f0a94a3137a6bc27/chromadb/segment/impl/vector/hnsw_params.py#L56-L59
	idxQuery := fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s_hsnwidx ON %s USING 
				hnsw (embedding vector_l2_ops)
				WITH (m = %d, ef_construction = %d);`,
		o.TableName,
		o.TableName,
		16,
		100,
	)
	if _, err := o.DB.Exec(idxQuery); err != nil {
		return nil, err
	}

	return &Store{
		pgvOptions: o,
	}, nil
}

func (p *Store) AddDocuments(ctx context.Context, documents []schema.Document, option ...vectorstores.Option) error {
	var contents []string
	for _, doc := range documents {
		contents = append(contents, doc.PageContent)
	}
	e, err := p.Embedder.EmbedDocuments(ctx, contents)
	if err != nil {
		return err
	}

	embeddingVectors := NewVectorSlice(e)

	txn, err := p.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn(p.TableName, "embedding", "metadata", "content"))
	if err != nil {
		return err
	}

	for i, doc := range documents {
		_, err = stmt.Exec(embeddingVectors[i], dbMap(doc.Metadata), doc.PageContent)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	if err = stmt.Close(); err != nil {
		return err
	}
	return txn.Commit()
}

func (p *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	txn, err := p.DB.Begin()
	if err != nil {
		return nil, err
	}

	// get query embeddings
	e, err := p.Embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	// query
	selectQuery := fmt.Sprintf(
		`SELECT content, metadata, embedding <-> $1 distance FROM %s 
				 ORDER BY distance LIMIT %d;`,
		p.TableName, numDocuments)
	rows, err := p.DB.Query(selectQuery, NewVector(e))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []schema.Document
	for rows.Next() {
		doc := schema.Document{}
		metadata := dbMap{}
		var distance float32
		if err = rows.Scan(&doc.PageContent, &metadata, &distance); err != nil {
			return nil, err
		}
		if metadata == nil {
			metadata = dbMap{}
		}
		doc.Metadata = metadata
		doc.Metadata["_query_distance"] = distance
		doc.Score = distance
		docs = append(docs, doc)
	}

	return docs, txn.Commit()
}

func (p *Store) CollectionSize() (int, error) {
	selectQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, p.TableName)
	row := p.DB.QueryRow(selectQuery)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, row.Err()
}

func (p *Store) CollectionName() string {
	return p.TableName
}

func (p *Store) ClearCollection() error {
	query := fmt.Sprintf(`TRUNCATE TABLE %s;`, p.TableName)
	_, err := p.DB.Exec(query)
	return err
}

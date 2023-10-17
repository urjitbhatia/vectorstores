package chroma

import (
	"fmt"
	"github.com/urjitbhatia/gochroma/embeddings"
)

type DistanceFn = string

var (
	DistanceFnL2 = "l2"
)

type chromaOptions struct {
	addr               string
	collectionName     string
	collectionMetadata map[string]any
	distanceFn         DistanceFn
	embedder           embeddings.Embedder
}
type Options func(o *chromaOptions)

// WithAddr Sets the address of the remote chroma instance. [Required]
func WithAddr(addr string) Options {
	return func(o *chromaOptions) {
		o.addr = addr
	}
}

// WithCollection Sets the collection name to target with this instance of the client. [Required]
// Can also optionally supply a metadata map to attach to the collection.
// To work with multiple collections, create a client per collection
func WithCollection(name string, metadata map[string]any) Options {
	return func(o *chromaOptions) {
		o.collectionName = name
		o.collectionMetadata = metadata
	}
}

// WithEmbedder Sets the embedder. [Required]
func WithEmbedder(e embeddings.Embedder) Options {
	return func(o *chromaOptions) {
		o.embedder = e
	}
}

// WithDistanceFn Sets the distance function. Defaults to "l2"/
// See Chroma docs for other supported distance functions
func WithDistanceFn(fn DistanceFn) Options {
	return func(o *chromaOptions) {
		o.distanceFn = fn
	}
}

func applyOptions(opts ...Options) (chromaOptions, error) {
	o := chromaOptions{distanceFn: DistanceFnL2}
	for _, fn := range opts {
		fn(&o)
	}
	// check required fields
	if o.addr == "" {
		return o, fmt.Errorf("missing required option: addr")
	}
	if o.collectionName == "" {
		return o, fmt.Errorf("missing required option: collectionName")
	}
	return o, nil
}

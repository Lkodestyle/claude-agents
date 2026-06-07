package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
)

const collectionName = "jarvis"

// Store wraps chromem-go behind domain-oriented ops.
// Keeping the chromem types out of handler code makes the eventual migration
// to Qdrant (phase 2) a one-file change.
//
// queryEmbed is optional and enables asymmetric retrieval: if the provider
// exposes separate document/query embeddings (Voyage does), we compute the
// query vector with queryEmbed and pass it to QueryEmbedding. When nil, we
// fall back to the collection's default embedding for both sides.
type Store struct {
	db         *chromem.DB
	coll       *chromem.Collection
	queryEmbed chromem.EmbeddingFunc
}

func OpenStore(dataDir string, docEmbed, queryEmbed chromem.EmbeddingFunc) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}

	db, err := chromem.NewPersistentDB(dataDir, false)
	if err != nil {
		return nil, fmt.Errorf("open persistent db: %w", err)
	}

	coll, err := db.GetOrCreateCollection(collectionName, nil, docEmbed)
	if err != nil {
		return nil, fmt.Errorf("open collection: %w", err)
	}

	return &Store{db: db, coll: coll, queryEmbed: queryEmbed}, nil
}

func (s *Store) Remember(ctx context.Context, m Memory) (string, error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	meta := map[string]string{
		"scope":      string(m.Scope),
		"tags":       strings.Join(m.Tags, ","),
		"source":     m.Source,
		"created_at": m.CreatedAt.Format(time.RFC3339),
	}

	doc := chromem.Document{
		ID:       m.ID,
		Metadata: meta,
		Content:  m.Content,
	}

	if err := s.coll.AddDocuments(ctx, []chromem.Document{doc}, runtime.NumCPU()); err != nil {
		return "", fmt.Errorf("add document: %w", err)
	}
	return m.ID, nil
}

func (s *Store) Recall(ctx context.Context, query string, k int, scope MemoryScope) ([]RecallHit, error) {
	if k <= 0 {
		k = 5
	}
	if k > 20 {
		k = 20
	}
	count := s.coll.Count()
	if count == 0 {
		return []RecallHit{}, nil
	}
	if k > count {
		k = count
	}

	var where map[string]string
	if scope != "" {
		where = map[string]string{"scope": string(scope)}
	}

	var (
		results []chromem.Result
		err     error
	)
	if s.queryEmbed != nil {
		qEmb, embErr := s.queryEmbed(ctx, query)
		if embErr != nil {
			return nil, fmt.Errorf("embed query: %w", embErr)
		}
		results, err = s.coll.QueryEmbedding(ctx, qEmb, k, where, nil)
	} else {
		results, err = s.coll.Query(ctx, query, k, where, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	hits := make([]RecallHit, 0, len(results))
	for _, r := range results {
		hits = append(hits, toHit(r))
	}
	return hits, nil
}

func (s *Store) Forget(ctx context.Context, id string) error {
	return s.coll.Delete(ctx, nil, nil, id)
}

func toHit(r chromem.Result) RecallHit {
	var tags []string
	if raw := r.Metadata["tags"]; raw != "" {
		tags = strings.Split(raw, ",")
	}
	return RecallHit{
		ID:         r.ID,
		Content:    r.Content,
		Similarity: r.Similarity,
		Scope:      MemoryScope(r.Metadata["scope"]),
		Tags:       tags,
		Source:     r.Metadata["source"],
		CreatedAt:  r.Metadata["created_at"],
	}
}

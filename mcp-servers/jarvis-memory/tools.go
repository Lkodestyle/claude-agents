package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server carries the deps shared by every tool handler.
type Server struct {
	store *Store
}

func (srv *Server) handleRemember(ctx context.Context, _ *mcp.CallToolRequest, in RememberInput) (*mcp.CallToolResult, RememberOutput, error) {
	if in.Content == "" {
		return nil, RememberOutput{}, fmt.Errorf("content is required")
	}
	if in.Scope == "" {
		in.Scope = ScopeSession
	}

	id, err := srv.store.Remember(ctx, Memory{
		Content: in.Content,
		Scope:   in.Scope,
		Tags:    in.Tags,
		Source:  in.Source,
	})
	if err != nil {
		return nil, RememberOutput{}, err
	}
	return nil, RememberOutput{ID: id}, nil
}

func (srv *Server) handleRecall(ctx context.Context, _ *mcp.CallToolRequest, in RecallInput) (*mcp.CallToolResult, RecallOutput, error) {
	if in.Query == "" {
		return nil, RecallOutput{}, fmt.Errorf("query is required")
	}
	hits, err := srv.store.Recall(ctx, in.Query, in.K, in.Scope)
	if err != nil {
		return nil, RecallOutput{}, err
	}
	return nil, RecallOutput{Hits: hits}, nil
}

func (srv *Server) handleForget(ctx context.Context, _ *mcp.CallToolRequest, in ForgetInput) (*mcp.CallToolResult, ForgetOutput, error) {
	if in.ID == "" {
		return nil, ForgetOutput{}, fmt.Errorf("id is required")
	}
	if err := srv.store.Forget(ctx, in.ID); err != nil {
		return nil, ForgetOutput{}, err
	}
	return nil, ForgetOutput{OK: true}, nil
}

// handleReflect samples top-K per scope and hands the synthesis back to Claude.
// We don't summarize here; producing coherent reflections is the model's job.
func (srv *Server) handleReflect(ctx context.Context, _ *mcp.CallToolRequest, in ReflectInput) (*mcp.CallToolResult, ReflectOutput, error) {
	k := in.K
	if k <= 0 {
		k = 5
	}
	topic := in.Topic
	if topic == "" {
		topic = "recent context"
	}

	out := ReflectOutput{ByScope: make(map[MemoryScope][]RecallHit)}
	for _, scope := range []MemoryScope{ScopeUser, ScopeFeedback, ScopeProject, ScopeReference} {
		hits, err := srv.store.Recall(ctx, topic, k, scope)
		if err != nil {
			return nil, out, fmt.Errorf("recall %s: %w", scope, err)
		}
		if len(hits) > 0 {
			out.ByScope[scope] = hits
		}
	}
	return nil, out, nil
}

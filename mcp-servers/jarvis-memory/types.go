package main

import "time"

// MemoryScope classifies a memory by its purpose. Mirrors the taxonomy used
// by the agent's file-based memory system (see CLAUDE.md), so recall results
// line up with how Jarvis already thinks about what to store.
type MemoryScope string

const (
	ScopeUser      MemoryScope = "user"      // stable facts about the person
	ScopeFeedback  MemoryScope = "feedback"  // guidance on how to collaborate
	ScopeProject   MemoryScope = "project"   // ongoing work context
	ScopeReference MemoryScope = "reference" // pointers to external systems
	ScopeSession   MemoryScope = "session"   // transient, conversation-scoped
)

// Memory is one stored unit of knowledge.
type Memory struct {
	ID        string      `json:"id"`
	Content   string      `json:"content"`
	Scope     MemoryScope `json:"scope"`
	Tags      []string    `json:"tags,omitempty"`
	Source    string      `json:"source,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// --- Tool I/O schemas (tags drive the JSON Schema inference in the MCP SDK) ---

type RememberInput struct {
	Content string      `json:"content" jsonschema:"the memory to store, as a self-contained statement"`
	Scope   MemoryScope `json:"scope" jsonschema:"one of: user, feedback, project, reference, session"`
	Tags    []string    `json:"tags,omitempty" jsonschema:"optional free-form tags for filtering"`
	Source  string      `json:"source,omitempty" jsonschema:"optional provenance (conversation id, file path, etc)"`
}

type RememberOutput struct {
	ID string `json:"id"`
}

type RecallInput struct {
	Query string      `json:"query" jsonschema:"natural-language question or topic to recall"`
	K     int         `json:"k,omitempty" jsonschema:"how many results to return (default 5, max 20)"`
	Scope MemoryScope `json:"scope,omitempty" jsonschema:"optional scope filter"`
}

type RecallHit struct {
	ID         string      `json:"id"`
	Content    string      `json:"content"`
	Similarity float32     `json:"similarity"`
	Scope      MemoryScope `json:"scope,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	Source     string      `json:"source,omitempty"`
	CreatedAt  string      `json:"created_at,omitempty"`
}

type RecallOutput struct {
	Hits []RecallHit `json:"hits"`
}

type ForgetInput struct {
	ID string `json:"id" jsonschema:"id of the memory to delete"`
}

type ForgetOutput struct {
	OK bool `json:"ok"`
}

type ReflectInput struct {
	Topic string `json:"topic,omitempty" jsonschema:"optional topic to focus the reflection on"`
	K     int    `json:"k,omitempty" jsonschema:"how many memories to sample per scope (default 5)"`
}

type ReflectOutput struct {
	ByScope map[MemoryScope][]RecallHit `json:"by_scope"`
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Brain is whatever turns a user utterance into spoken replies. It is
// event-driven so a single turn can produce several things to say (a quick
// "dale, reviso" before a tool call, then the result) and can pause to ask
// for a spoken permission confirmation.
type Brain interface {
	// Send starts a turn with the user's words. Events for the turn arrive
	// on Events() and always end with an EventDone.
	Send(text string) error
	Events() <-chan BrainEvent
	// Respond answers an EventPermission.
	Respond(requestID string, allow bool, reason string) error
	// Interrupt aborts the running turn (best effort).
	Interrupt() error
	Close() error
}

type BrainEventKind int

const (
	EventText       BrainEventKind = iota // something to say
	EventTool                             // a tool started (for the console)
	EventPermission                       // needs a yes/no from the user
	EventDone                             // turn finished
	EventError                            // turn failed; Text holds the reason
)

type BrainEvent struct {
	Kind      BrainEventKind
	Text      string
	RequestID string          // EventPermission only
	Prompt    string          // EventPermission only: human description of the action
	ToolName  string          // EventPermission only
	Input     json.RawMessage // EventPermission only: raw tool input
}

// apiBrain wraps the original tool-less Messages API chat (plus optional
// jarvis-memory recall/remember) behind the Brain interface.
type apiBrain struct {
	cfg     *Config
	mem     *MemoryClient
	history []Message
	events  chan BrainEvent
	cancel  context.CancelFunc
}

func newAPIBrain(cfg *Config, mem *MemoryClient) *apiBrain {
	return &apiBrain{cfg: cfg, mem: mem, events: make(chan BrainEvent, 16)}
}

func (b *apiBrain) Events() <-chan BrainEvent { return b.events }

func (b *apiBrain) Respond(string, bool, string) error { return nil }

func (b *apiBrain) Interrupt() error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *apiBrain) Close() error { return b.Interrupt() }

func (b *apiBrain) Send(text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	b.cancel = cancel
	b.history = append(b.history, Message{Role: "user", Content: text})
	history := append([]Message(nil), b.history...)

	go func() {
		defer cancel()
		reply, err := b.chat(ctx, text, history)
		if err != nil {
			// Drop the unanswered user turn so history stays alternating.
			b.history = b.history[:len(b.history)-1]
			b.events <- BrainEvent{Kind: EventError, Text: err.Error()}
			b.events <- BrainEvent{Kind: EventDone}
			return
		}
		b.history = trimHistory(append(b.history, Message{Role: "assistant", Content: reply}), historyMax)
		b.events <- BrainEvent{Kind: EventText, Text: reply}
		b.events <- BrainEvent{Kind: EventDone}
	}()
	return nil
}

func (b *apiBrain) chat(ctx context.Context, transcript string, history []Message) (string, error) {
	// Memory recall is best-effort: a slow or broken memory server must not
	// stall the voice loop, so failures just log and the turn proceeds.
	system := b.cfg.SystemPrompt
	if b.mem != nil {
		recallCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		hits, err := b.mem.Recall(recallCtx, transcript, recallK, recallMinSimilarity)
		cancel()
		if err != nil {
			log.Printf("memory recall: %v", err)
		} else if len(hits) > 0 {
			fmt.Printf("• remembering (%d hits)...\n", len(hits))
			system += memoryContextBlock(hits)
		}
	}

	reply, err := anthropicChat(ctx, b.cfg, system, history)
	if err != nil {
		return "", fmt.Errorf("llm: %w", err)
	}

	// Persist the exchange in the background; the turn shouldn't wait on it.
	if b.mem != nil {
		exchange := fmt.Sprintf("El usuario dijo: %s\nJarvis respondio: %s", transcript, reply)
		go func() {
			memCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := b.mem.Remember(memCtx, exchange); err != nil {
				log.Printf("memory remember: %v", err)
			}
		}()
	}
	return reply, nil
}

// trimHistory keeps the latest N messages, preserving the user/assistant pair
// boundary so we don't ship Claude an orphaned assistant turn.
func trimHistory(h []Message, max int) []Message {
	if len(h) <= max {
		return h
	}
	start := len(h) - max
	// Make sure we don't start on an assistant turn (Anthropic requires the
	// first message to be from the user).
	if h[start].Role != "user" && start+1 < len(h) {
		start++
	}
	return h[start:]
}

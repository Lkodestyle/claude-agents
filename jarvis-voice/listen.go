package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// Listener gives hands-free turn taking: the mic is captured continuously,
// a cheap local energy gate decides when someone is talking, and only those
// stretches are streamed to Deepgram's live endpoint. Deepgram detects the
// end of the utterance, and the finished sentence comes out of Utterances().
//
// The gate matters for cost: Deepgram bills streamed audio, so silence is
// never sent — the socket is kept alive with KeepAlive messages instead.
//
// While Jarvis is speaking the listener is muted (half duplex), otherwise it
// would transcribe its own voice coming out of the speakers.
type Listener struct {
	cfg       *Config
	threshold float64
	out       chan string

	muted  atomic.Bool // set by the speaker during playback
	paused atomic.Bool // toggled by the user (Enter)

	connMu sync.Mutex
	conn   *websocket.Conn

	textMu sync.Mutex
	finals []string // is_final segments of the current utterance
}

const (
	sampleRate     = 16000
	frameMs        = 20
	frameBytes     = sampleRate * 2 * frameMs / 1000 // s16le mono
	prerollFrames  = 15                              // 300ms kept before the gate opens
	openFrames     = 3                               // 60ms above threshold to open
	keepAliveEvery = 5 * time.Second                 // Deepgram drops idle sockets at ~10s
)

func NewListener(cfg *Config) *Listener {
	return &Listener{cfg: cfg, out: make(chan string, 8)}
}

func (l *Listener) Utterances() <-chan string { return l.out }

// gateHangover is how long the gate stays open after the last loud frame.
// It outlasts Deepgram's utterance-end pause so the turn is decided there,
// not cut short by our Finalize.
func (l *Listener) gateHangover() time.Duration {
	return time.Duration(l.cfg.PauseMs+500) * time.Millisecond
}

// SetMuted is called by the speaker around playback.
func (l *Listener) SetMuted(m bool) { l.muted.Store(m) }

// TogglePause flips user pause and returns the new state.
func (l *Listener) TogglePause() bool {
	p := !l.paused.Load()
	l.paused.Store(p)
	return p
}

// Run captures audio until ctx is cancelled.
func (l *Listener) Run(ctx context.Context) error {
	args, err := captureArgs()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}
	defer cmd.Wait()
	defer l.closeConn()

	frame := make([]byte, frameBytes)
	read := func() ([]byte, error) {
		if _, err := io.ReadFull(stdout, frame); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("mic capture ended: %w\n%s", err, stderr.String())
		}
		return frame, nil
	}

	if err := l.calibrate(read); err != nil {
		return err
	}

	var (
		open     bool
		loud     int
		lastLoud time.Time
		lastSend = time.Now()
		preroll  [][]byte
	)
	closeGate := func() {
		open, loud, preroll = false, 0, preroll[:0]
		// Flush whatever Deepgram still holds for this utterance.
		l.sendJSON(ctx, map[string]string{"type": "Finalize"})
	}

	for {
		f, err := read()
		if err != nil {
			return err
		}
		now := time.Now()

		if l.muted.Load() || l.paused.Load() {
			if open {
				closeGate()
			}
			preroll = preroll[:0]
			if now.Sub(lastSend) > keepAliveEvery {
				l.keepAlive(ctx)
				lastSend = now
			}
			continue
		}

		level := rms(f)
		if !open {
			preroll = append(preroll, append([]byte(nil), f...))
			if len(preroll) > prerollFrames {
				preroll = preroll[1:]
			}
			if level >= l.threshold {
				loud++
			} else {
				loud = 0
			}
			if loud >= openFrames {
				open, lastLoud = true, now
				for _, p := range preroll {
					l.sendAudio(ctx, p)
				}
				preroll = preroll[:0]
				lastSend = now
			} else if now.Sub(lastSend) > keepAliveEvery {
				l.keepAlive(ctx)
				lastSend = now
			}
			continue
		}

		l.sendAudio(ctx, f)
		lastSend = now
		if level >= l.threshold {
			lastLoud = now
		} else if now.Sub(lastLoud) > l.gateHangover() {
			closeGate()
		}
	}
}

// calibrate measures ambient noise for one second and sets the gate
// threshold above it, unless the user fixed one via JARVIS_VOICE_VAD_RMS.
func (l *Listener) calibrate(read func() ([]byte, error)) error {
	switch {
	case l.cfg.VADThreshold > 0:
		l.threshold = float64(l.cfg.VADThreshold)
		return nil
	case l.cfg.VADThreshold < 0:
		l.threshold = 0 // gate always open
		return nil
	}

	fmt.Print("  calibrating mic (stay quiet 1s)... ")
	var sum float64
	n := 1000 / frameMs
	for i := 0; i < n; i++ {
		f, err := read()
		if err != nil {
			return err
		}
		sum += rms(f)
	}
	floor := sum / float64(n)
	l.threshold = math.Max(floor*3, 300)
	fmt.Printf("noise %.0f → gate %.0f\n", floor, l.threshold)
	return nil
}

func rms(frame []byte) float64 {
	var sum float64
	n := len(frame) / 2
	for i := 0; i < n; i++ {
		s := float64(int16(binary.LittleEndian.Uint16(frame[2*i:])))
		sum += s * s
	}
	return math.Sqrt(sum / float64(n))
}

// --- Deepgram live connection ---

func (l *Listener) dial(ctx context.Context) (*websocket.Conn, error) {
	q := url.Values{}
	q.Set("model", "nova-3")
	q.Set("language", getenv("JARVIS_VOICE_LANG", "multi"))
	q.Set("encoding", "linear16")
	q.Set("sample_rate", fmt.Sprint(sampleRate))
	q.Set("channels", "1")
	q.Set("smart_format", "true")
	q.Set("punctuate", "true")
	q.Set("interim_results", "true") // required for UtteranceEnd
	q.Set("endpointing", "400")
	q.Set("utterance_end_ms", fmt.Sprint(l.cfg.PauseMs))

	h := http.Header{}
	h.Set("Authorization", "Token "+l.cfg.DeepgramKey)
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(dialCtx, "wss://api.deepgram.com/v1/listen?"+q.Encode(),
		&websocket.DialOptions{HTTPHeader: h})
	if err != nil {
		return nil, fmt.Errorf("dial deepgram: %w", err)
	}
	c.SetReadLimit(1 << 20)
	go l.readResults(ctx, c)
	return c, nil
}

// getConn returns the live socket, dialing on demand. Holding connMu while
// dialing is fine: only the capture loop writes.
func (l *Listener) getConn(ctx context.Context) *websocket.Conn {
	l.connMu.Lock()
	defer l.connMu.Unlock()
	if l.conn != nil {
		return l.conn
	}
	c, err := l.dial(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("stt: %v", err)
		}
		return nil
	}
	l.conn = c
	return c
}

func (l *Listener) dropConn(c *websocket.Conn) {
	l.connMu.Lock()
	if l.conn == c {
		l.conn = nil
	}
	l.connMu.Unlock()
	c.CloseNow()
}

func (l *Listener) closeConn() {
	l.connMu.Lock()
	c := l.conn
	l.conn = nil
	l.connMu.Unlock()
	if c != nil {
		c.Write(context.Background(), websocket.MessageText, []byte(`{"type":"CloseStream"}`))
		c.CloseNow()
	}
}

func (l *Listener) sendAudio(ctx context.Context, frame []byte) {
	if c := l.getConn(ctx); c != nil {
		if err := c.Write(ctx, websocket.MessageBinary, frame); err != nil {
			l.dropConn(c)
		}
	}
}

func (l *Listener) sendJSON(ctx context.Context, v any) {
	l.connMu.Lock()
	c := l.conn
	l.connMu.Unlock()
	if c == nil {
		return // nothing to finalize or keep alive
	}
	data, _ := json.Marshal(v)
	if err := c.Write(ctx, websocket.MessageText, data); err != nil {
		l.dropConn(c)
	}
}

func (l *Listener) keepAlive(ctx context.Context) {
	l.sendJSON(ctx, map[string]string{"type": "KeepAlive"})
}

type dgMessage struct {
	Type    string `json:"type"`
	IsFinal bool   `json:"is_final"`
	Channel struct {
		Alternatives []struct {
			Transcript string `json:"transcript"`
		} `json:"alternatives"`
	} `json:"channel"`
	FromFinalize bool `json:"from_finalize"`
}

func (l *Listener) readResults(ctx context.Context, c *websocket.Conn) {
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			var ce websocket.CloseError
			if ctx.Err() == nil && !errors.As(err, &ce) && !errors.Is(err, net.ErrClosed) {
				log.Printf("stt read: %v", err)
			}
			l.dropConn(c)
			return
		}
		var m dgMessage
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		switch m.Type {
		case "Results":
			if m.IsFinal && len(m.Channel.Alternatives) > 0 {
				if t := strings.TrimSpace(m.Channel.Alternatives[0].Transcript); t != "" {
					l.textMu.Lock()
					l.finals = append(l.finals, t)
					l.textMu.Unlock()
				}
			}
			if m.FromFinalize {
				l.flush()
			}
		case "UtteranceEnd":
			l.flush()
		}
	}
}

// flush emits the accumulated utterance, if any.
func (l *Listener) flush() {
	l.textMu.Lock()
	text := strings.Join(l.finals, " ")
	l.finals = nil
	l.textMu.Unlock()
	if text != "" {
		l.out <- text
	}
}

// captureArgs builds an ffmpeg command line that streams raw 16kHz mono
// s16le PCM from the mic to stdout. JARVIS_VOICE_FFMPEG_INPUT replaces the
// input part entirely (e.g. "-re -i test.wav" to replay a recording).
func captureArgs() ([]string, error) {
	var in []string
	if raw := os.Getenv("JARVIS_VOICE_FFMPEG_INPUT"); raw != "" {
		in = strings.Fields(raw)
	} else {
		dev, err := ResolveInputDevice()
		if err != nil {
			return nil, err
		}
		in = []string{"-f", inputFormat(), "-i", dev}
	}
	args := append([]string{"-loglevel", "error", "-hide_banner"}, in...)
	return append(args, "-ac", "1", "-ar", fmt.Sprint(sampleRate), "-f", "s16le", "-"), nil
}

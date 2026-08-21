package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// Edge TTS (Microsoft) — the same engine the Edge browser uses for "Read
// Aloud". It's not officially exposed as a public API, but the protocol is
// stable and widely consumed by community tools (python edge-tts, etc.).
//
// We implement it directly because:
//   1. It's free and unmetered.
//   2. The wire format is small (one WSS connection, two text frames out,
//      a stream of binary frames in).
//   3. Avoids pulling a heavy SDK dependency.
//
// Auth scheme requires a synthetic Sec-MS-GEC header — a SHA256 hash of a
// time-bucketed value plus a static client token. The token is shared with
// every Edge install on earth; this is not a credential, just a soft gate.

const (
	edgeTrustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	edgeTTSURL             = "wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=" + edgeTrustedClientToken
	edgeGECVersion         = "1-131.0.2903.86"
	edgeUserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"
	edgeOrigin             = "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold"
)

func edgeTTS(ctx context.Context, voice, text, outPath string) error {
	headers := http.Header{}
	headers.Set("Sec-MS-GEC", generateSecMsGec())
	headers.Set("Sec-MS-GEC-Version", edgeGECVersion)
	headers.Set("User-Agent", edgeUserAgent)
	headers.Set("Origin", edgeOrigin)

	c, _, err := websocket.Dial(ctx, edgeTTSURL, &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return fmt.Errorf("dial edge-tts: %w", err)
	}
	defer c.CloseNow()

	// Audio chunks can be sizeable; default 32KB read limit is too small.
	c.SetReadLimit(16 * 1024 * 1024)

	// 1) speech.config — tells the server what audio format we want.
	timestamp := time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT+0000 (Coordinated Universal Time)")
	cfgMsg := "X-Timestamp:" + timestamp + "\r\n" +
		"Content-Type:application/json; charset=utf-8\r\n" +
		"Path:speech.config\r\n\r\n" +
		`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"false"},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`
	if err := c.Write(ctx, websocket.MessageText, []byte(cfgMsg)); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// 2) ssml — the actual text to speak, wrapped in SSML.
	reqID := strings.ReplaceAll(uuid.NewString(), "-", "")
	ssml := fmt.Sprintf(
		`<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'><voice name='%s'><prosody pitch='+0Hz' rate='+0%%' volume='+0%%'>%s</prosody></voice></speak>`,
		voice, escapeXML(text),
	)
	ssmlMsg := "X-RequestId:" + reqID + "\r\n" +
		"Content-Type:application/ssml+xml\r\n" +
		"X-Timestamp:" + timestamp + "\r\n" +
		"Path:ssml\r\n\r\n" + ssml
	if err := c.Write(ctx, websocket.MessageText, []byte(ssmlMsg)); err != nil {
		return fmt.Errorf("write ssml: %w", err)
	}

	// 3) Stream binary frames until we get a turn.end text frame.
	//    Each binary frame layout:
	//      [0..1]              uint16 BE: header length
	//      [2..2+headerLen]    text headers
	//      [2+headerLen..]     audio bytes
	var buf bytes.Buffer
	for {
		typ, data, err := c.Read(ctx)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		switch typ {
		case websocket.MessageBinary:
			if len(data) < 2 {
				continue
			}
			headerLen := int(binary.BigEndian.Uint16(data[:2]))
			if 2+headerLen > len(data) {
				continue
			}
			buf.Write(data[2+headerLen:])
		case websocket.MessageText:
			if strings.Contains(string(data), "Path:turn.end") {
				if buf.Len() == 0 {
					return fmt.Errorf("edge-tts returned no audio")
				}
				return os.WriteFile(outPath, buf.Bytes(), 0o644)
			}
		}
	}
}

// generateSecMsGec produces the Sec-MS-GEC header value Microsoft expects.
// Python reference: https://github.com/rany2/edge-tts (drainage_token / sec_ms_gec).
//
// Quirk: the timestamp is in 100-ns ticks since the Windows epoch (1601-01-01),
// rounded down to the nearest 5-minute boundary so the same value holds for a
// few minutes — this lets the server cache the validation cheaply.
func generateSecMsGec() string {
	const winEpoch = int64(11644473600)         // seconds between 1601 and 1970
	const interval = int64(5 * 60 * 10_000_000) // 5 minutes in 100-ns ticks
	ticks := (time.Now().Unix() + winEpoch) * 10_000_000
	ticks -= ticks % interval
	h := sha256.Sum256([]byte(fmt.Sprintf("%d%s", ticks, edgeTrustedClientToken)))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

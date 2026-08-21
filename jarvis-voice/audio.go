package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// inputDeviceCache holds the resolved ffmpeg input arg (e.g. "audio=Micrófono (HyperX Cloud III)")
// after the first successful resolution, so we don't re-scan devices each turn.
var inputDeviceCache string

// recordWAV captures audio from the input device and writes a WAV file at
// outPath. Uses ffmpeg as a subprocess; choice driven by:
//   - cross-platform (DirectShow on Windows, AVFoundation on macOS, ALSA on Linux)
//   - no native deps in our binary (pure-Go promise stays intact)
//   - excellent format support, clean exit on context cancel
//
// 16kHz mono is what Deepgram expects for the cheapest tier.
func recordWAV(ctx context.Context, outPath string, dur time.Duration) error {
	args, err := platformRecordArgs(dur)
	if err != nil {
		return err
	}
	args = append(args,
		"-ac", "1",
		"-ar", "16000",
		"-y", // overwrite
		outPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg record failed: %w\n%s", err, string(out))
	}
	return nil
}

// ResolveInputDevice picks an ffmpeg input spec (e.g. "audio=Micrófono (...)")
// for the current OS. Override priority:
//   1. JARVIS_VOICE_INPUT env var — full ffmpeg input spec, used as-is.
//   2. JARVIS_VOICE_INPUT_NAME env var — substring match against detected
//      audio device names (case-insensitive).
//   3. Auto-detect: enumerate audio devices, prefer known headset brands.
//   4. Fall back to OS-typical defaults.
//
// Result is cached for the lifetime of the process — no need to re-scan
// every turn.
func ResolveInputDevice() (string, error) {
	if v := os.Getenv("JARVIS_VOICE_INPUT"); v != "" {
		return v, nil
	}
	if inputDeviceCache != "" {
		return inputDeviceCache, nil
	}

	switch runtime.GOOS {
	case "windows":
		dev, err := detectWindowsAudioInput()
		if err != nil {
			return "", err
		}
		inputDeviceCache = dev
		return dev, nil
	case "darwin":
		inputDeviceCache = ":0"
		return inputDeviceCache, nil
	default:
		inputDeviceCache = "default"
		return inputDeviceCache, nil
	}
}

func platformRecordArgs(dur time.Duration) ([]string, error) {
	secs := strconv.Itoa(int(dur.Seconds()))
	input, err := ResolveInputDevice()
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "windows":
		return []string{
			"-loglevel", "error",
			"-f", "dshow",
			"-i", input,
			"-t", secs,
		}, nil
	case "darwin":
		return []string{
			"-loglevel", "error",
			"-f", "avfoundation",
			"-i", input,
			"-t", secs,
		}, nil
	default:
		return []string{
			"-loglevel", "error",
			"-f", "alsa",
			"-i", input,
			"-t", secs,
		}, nil
	}
}

// detectWindowsAudioInput parses `ffmpeg -f dshow -list_devices true -i dummy`
// and picks an audio device. We honor JARVIS_VOICE_INPUT_NAME as a substring
// preference; otherwise we score devices and pick the most likely real mic.
func detectWindowsAudioInput() (string, error) {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-f", "dshow", "-list_devices", "true", "-i", "dummy")
	out, _ := cmd.CombinedOutput() // exits 1 because "dummy" isn't a real device; output still useful

	re := regexp.MustCompile(`"([^"]+)"\s+\(audio\)`)
	matches := re.FindAllStringSubmatch(string(out), -1)
	if len(matches) == 0 {
		return "", errors.New("no audio devices found via dshow list_devices")
	}

	devices := make([]string, 0, len(matches))
	for _, m := range matches {
		devices = append(devices, m[1])
	}

	// Explicit user preference wins.
	if want := strings.ToLower(os.Getenv("JARVIS_VOICE_INPUT_NAME")); want != "" {
		for _, d := range devices {
			if strings.Contains(strings.ToLower(d), want) {
				return "audio=" + d, nil
			}
		}
	}

	// Heuristic: prefer known headset/mic brands over webcam-bundled mics.
	preferred := []string{
		"hyperx", "shure", "rode", "rØde", "blue yeti", "audio-technica",
		"sennheiser", "logitech g", "razer", "elgato", "samson",
	}
	for _, p := range preferred {
		for _, d := range devices {
			if strings.Contains(strings.ToLower(d), p) {
				return "audio=" + d, nil
			}
		}
	}

	// Otherwise: skip webcam-named devices, take the first plain mic.
	for _, d := range devices {
		ld := strings.ToLower(d)
		if strings.Contains(ld, "webcam") || strings.Contains(ld, "camera") {
			continue
		}
		return "audio=" + d, nil
	}

	// Last resort.
	return "audio=" + devices[0], nil
}

// playMP3 blocks until ffplay finishes playing inPath. We pick ffplay over
// ffmpeg+pipe because it handles audio output and OS quirks for us, and we
// don't need streaming for a 5-15 second reply.
func playMP3(ctx context.Context, inPath string) error {
	if _, err := exec.LookPath("ffplay"); err != nil {
		return errors.New("ffplay not in PATH (install ffmpeg full build)")
	}
	cmd := exec.CommandContext(ctx, "ffplay",
		"-loglevel", "error",
		"-nodisp",   // no GUI window
		"-autoexit", // quit when file ends
		"-hide_banner",
		inPath,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffplay failed: %w", err)
	}
	return nil
}

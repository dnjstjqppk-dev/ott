package service

import (
	"bytes"
	"strings"
	"testing"

	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
)

func TestCountCueRunesAndWebVttFormat(t *testing.T) {
	lines := []string{"안녕하세요 넷플릭스", "오리지널 시리즈입니다"}
	total, maxLine := countCueRunes(lines)
	if total != 21 || maxLine != 11 {
		t.Fatalf("unexpected rune counts: total=%d maxLine=%d", total, maxLine)
	}

	ts := formatVttTime(3661500)
	if ts != "01:01:01.500" {
		t.Fatalf("expected 01:01:01.500, got %s", ts)
	}

	buf := bytes.NewBufferString("WEBVTT\n\n")
	writeVttCueBlock(buf, &ingestv1.TimedTextCueDto{
		CueSequence: 1,
		StartMs:     1500,
		EndMs:       4000,
		TextLines:   lines,
	})
	output := buf.String()
	if !strings.Contains(output, "00:00:01.500 --> 00:00:04.000") {
		t.Fatalf("missing expected cue timecode range in output: %s", output)
	}
}


package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseValidBasic(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n0-1\n"
	path := writeTempFile(t, content)
	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Graph.AntCount != 3 {
		t.Fatalf("expected 3 ants, got %d", result.Graph.AntCount)
	}
	if result.Graph.Start != "0" {
		t.Fatalf("expected start '0', got %q", result.Graph.Start)
	}
	if result.Graph.End != "1" {
		t.Fatalf("expected end '1', got %q", result.Graph.End)
	}
}

func TestParseNoAnts(t *testing.T) {
	content := "##start\n0 0 0\n##end\n1 10 0\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing ants")
	}
}

func TestParseInvalidAnts(t *testing.T) {
	content := "abc\n##start\n0 0 0\n##end\n1 10 0\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid ant count")
	}
}

func TestParseNoStart(t *testing.T) {
	content := "3\n##end\n1 10 0\n0 0 0\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing start")
	}
}

func TestParseNoEnd(t *testing.T) {
	content := "3\n##start\n0 0 0\n1 10 0\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing end")
	}
}

func TestParseNoPath(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n2 5 5\n0-2\n"
	path := writeTempFile(t, content)
	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("parse should succeed, got: %v", err)
	}
	if result.Graph.AntCount != 3 {
		t.Fatalf("expected 3 ants")
	}
}

func TestParseDuplicateRoom(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n0 5 5\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for duplicate room")
	}
}

func TestParseInvalidLinkRoom(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n0-2\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for unknown room in link")
	}
}

func TestParseSelfLink(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n0-0\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for self-link")
	}
}

func TestParseComments(t *testing.T) {
	content := "3\n#this is a comment\n##start\n0 0 0\n##end\n1 10 0\n#another comment\n0-1\n"
	path := writeTempFile(t, content)
	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Graph.AntCount != 3 {
		t.Fatalf("expected 3 ants")
	}
}

func TestParseRoomNameStartsWithL(t *testing.T) {
	content := "3\n##start\nLroom 0 0\n##end\n1 10 0\nLroom-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for room name starting with L")
	}
}

func TestParseHyphenRoomNames(t *testing.T) {
	content := "3\n##start\nstart-room 0 0\n##end\nend-room 10 0\nmiddle-room 5 5\nstart-room-middle-room\nmiddle-room-end-room\n"
	path := writeTempFile(t, content)
	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Graph.Start != "start-room" {
		t.Fatalf("expected start-room, got %q", result.Graph.Start)
	}
	if result.Graph.End != "end-room" {
		t.Fatalf("expected end-room, got %q", result.Graph.End)
	}
	if !result.Graph.HasLink("start-room", "middle-room") {
		t.Fatal("expected link start-room-middle-room")
	}
	if !result.Graph.HasLink("middle-room", "end-room") {
		t.Fatal("expected link middle-room-end-room")
	}
}

func TestParseNegativeCoordinates(t *testing.T) {
	content := "1\n##start\n0 0 0\n##end\n1 10 0\n2 -5 3\n0-2\n2-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseUnknownCommand(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n##unknown\n0-1\n"
	path := writeTempFile(t, content)
	_, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseLinesPreserved(t *testing.T) {
	content := "3\n##start\n0 0 0\n##end\n1 10 0\n#comment\n0-1\n"
	path := writeTempFile(t, content)
	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasStart := false
	hasEnd := false
	hasRoom0 := false
	hasRoom1 := false
	hasLink := false
	hasComment := false
	for _, line := range result.Lines {
		switch line {
		case "##start":
			hasStart = true
		case "##end":
			hasEnd = true
		case "0 0 0":
			hasRoom0 = true
		case "1 10 0":
			hasRoom1 = true
		case "0-1":
			hasLink = true
		case "#comment":
			hasComment = true
		}
	}
	if !hasStart {
		t.Fatal("expected ##start in preserved lines")
	}
	if !hasEnd {
		t.Fatal("expected ##end in preserved lines")
	}
	if !hasRoom0 {
		t.Fatal("expected room '0 0 0' in preserved lines")
	}
	if !hasRoom1 {
		t.Fatal("expected room '1 10 0' in preserved lines")
	}
	if !hasLink {
		t.Fatal("expected link '0-1' in preserved lines")
	}
	if hasComment {
		t.Fatal("did not expect #comment in preserved lines")
	}
}

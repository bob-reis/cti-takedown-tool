package models

import "testing"

func TestEvidencePack_GetDefangedURL(t *testing.T) {
	e := &EvidencePack{}
	original := "https://example.com/path"
	expected := "hxxps://example[.]com/path"
	if got := e.GetDefangedURL(original); got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
	// When Defanged preset
	e.Defanged = "custom"
	if got := e.GetDefangedURL("http://example.com"); got != "custom" {
		t.Fatalf("expected preset defanged value, got %s", got)
	}
}

func TestReplaceInString(t *testing.T) {
	s := replaceInString("foo bar foo", "foo", "baz")
	if s != "baz bar baz" {
		t.Fatalf("unexpected result: %s", s)
	}
}

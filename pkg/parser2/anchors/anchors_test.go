package anchors

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml/ast"
)

func TestLoadAnchorsFromDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors")
	os.MkdirAll(dir, 0755)
	yamlContent := `
foo: &myanchor
  bar: baz
`
	filePath := filepath.Join(dir, "test.yaml")
	os.WriteFile(filePath, []byte(yamlContent), 0644)
	defer os.RemoveAll(dir)

	anchors, err := LoadAnchorsFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(anchors) != 1 {
		t.Fatalf("expected 1 anchor, got %d", len(anchors))
	}

	val, ok := anchors["myanchor"]
	if !ok {
		t.Fatalf("anchor 'myanchor' not found")
	}

	mapping, ok := val.(*ast.MappingNode)
	if !ok {
		t.Fatalf("anchor value is not a MappingNode")
	}

	if len(mapping.Values) != 1 {
		t.Fatalf("expected 1 mapping value, got %d", len(mapping.Values))
	}

	if mapping.Values[0].Key.String() != "bar" || mapping.Values[0].Value.String() != "baz" {
		t.Errorf("unexpected mapping content: got %s: %s", mapping.Values[0].Key.String(), mapping.Values[0].Value.String())
	}
}

func TestLoadAnchorsFromDir_MultipleFiles(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_multi")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml1 := `foo: &a
  bar: baz`
	yaml2 := `baz: &b
  qux: quux`
	f1 := filepath.Join(dir, "f1.yaml")
	f2 := filepath.Join(dir, "f2.yaml")
	os.WriteFile(f1, []byte(yaml1), 0644)
	os.WriteFile(f2, []byte(yaml2), 0644)

	anchors, err := LoadAnchorsFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anchors) != 2 {
		t.Fatalf("expected 2 anchors, got %d", len(anchors))
	}
	if _, ok := anchors["a"]; !ok {
		t.Errorf("anchor 'a' not found")
	}
	if _, ok := anchors["b"]; !ok {
		t.Errorf("anchor 'b' not found")
	}
}

func TestLoadAnchorsFromDir_AnchorUpdateInOneDocument(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_update")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml := `foo: &a
  bar: baz
foo2: &a
  bar: updated`
	f := filepath.Join(dir, "f.yaml")
	os.WriteFile(f, []byte(yaml), 0644)

	_, err := LoadAnchorsFromDir(dir)
	if err == nil {
		t.Fatalf("expected error for duplicate anchor, got nil")
	}
}

func TestLoadAnchorsFromDir_MultipleDocumentsInFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_docs")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml := `---
foo: &a
  bar: baz
---
bar: &b
  qux: quux`
	f := filepath.Join(dir, "f.yaml")
	os.WriteFile(f, []byte(yaml), 0644)

	anchors, err := LoadAnchorsFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anchors) != 2 {
		t.Fatalf("expected 2 anchors, got %d", len(anchors))
	}
	if _, ok := anchors["a"]; !ok {
		t.Errorf("anchor 'a' not found")
	}
	if _, ok := anchors["b"]; !ok {
		t.Errorf("anchor 'b' not found")
	}
}

func TestLoadAnchorsFromDir_DuplicateAcrossFiles(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_dupe")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml1 := `foo: &a
  bar: baz`
	yaml2 := `baz: &a
  qux: quux`
	f1 := filepath.Join(dir, "f1.yaml")
	f2 := filepath.Join(dir, "f2.yaml")
	os.WriteFile(f1, []byte(yaml1), 0644)
	os.WriteFile(f2, []byte(yaml2), 0644)

	_, err := LoadAnchorsFromDir(dir)
	if err == nil {
		t.Fatalf("expected error for duplicate anchor across files, got nil")
	}
}

func TestLoadAnchorsFromDir_DeeplyNestedAnchors(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_deep")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml := `root:
  level1:
    level2:
      level3: &deepanchor
        key: value
  sibling:
    foo: bar`
	f := filepath.Join(dir, "deep.yaml")
	os.WriteFile(f, []byte(yaml), 0644)

	anchors, err := LoadAnchorsFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anchors) != 1 {
		t.Fatalf("expected 1 anchor, got %d", len(anchors))
	}
	val, ok := anchors["deepanchor"]
	if !ok {
		t.Fatalf("anchor 'deepanchor' not found")
	}
	mapping, ok := val.(*ast.MappingNode)
	if !ok {
		t.Fatalf("anchor value is not a MappingNode")
	}
	if len(mapping.Values) != 1 {
		t.Fatalf("expected 1 mapping value, got %d", len(mapping.Values))
	}
	if mapping.Values[0].Key.String() != "key" || mapping.Values[0].Value.String() != "value" {
		t.Errorf("unexpected mapping content: got %s: %s", mapping.Values[0].Key.String(), mapping.Values[0].Value.String())
	}
}

func TestLoadAnchorsFromDir_AnchorInSequence(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "testanchors_seq")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	yaml := `seq:
  - &seqanchor
    foo: bar
  - baz: qux`
	f := filepath.Join(dir, "seq.yaml")
	os.WriteFile(f, []byte(yaml), 0644)

	anchors, err := LoadAnchorsFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anchors) != 1 {
		t.Fatalf("expected 1 anchor, got %d", len(anchors))
	}
	val, ok := anchors["seqanchor"]
	if !ok {
		t.Fatalf("anchor 'seqanchor' not found")
	}
	mapping, ok := val.(*ast.MappingNode)
	if !ok {
		t.Fatalf("anchor value is not a MappingNode")
	}
	if len(mapping.Values) != 1 {
		t.Fatalf("expected 1 mapping value, got %d", len(mapping.Values))
	}
	if mapping.Values[0].Key.String() != "foo" || mapping.Values[0].Value.String() != "bar" {
		t.Errorf("unexpected mapping content: got %s: %s", mapping.Values[0].Key.String(), mapping.Values[0].Value.String())
	}
}

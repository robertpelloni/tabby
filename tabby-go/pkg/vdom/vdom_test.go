package vdom

import (
	"encoding/json"
	"testing"
)

func TestVDOM_JSON(t *testing.T) {
	node := NewNode("div")
	node.SetProp("className", "p-4")

	child := NewNode("h1")
	child.AddChild("Title")

	node.AddChild(child)
	node.AddChild("Text")

	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"tag":"div","props":{"className":"p-4"},"children":[{"tag":"h1","children":["Title"]},"Text"]}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

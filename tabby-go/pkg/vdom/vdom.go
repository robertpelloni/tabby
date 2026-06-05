package vdom

type Node struct {
	Tag      string            `json:"tag"`
	Props    map[string]any    `json:"props,omitempty"`
	Children []any             `json:"children,omitempty"` // Can be *Node or string
}

type Tree struct {
	Root *Node `json:"root"`
}

func NewNode(tag string) *Node {
	return &Node{
		Tag:   tag,
		Props: make(map[string]any),
	}
}

func (n *Node) AddChild(child any) {
	n.Children = append(n.Children, child)
}

func (n *Node) SetProp(key string, value any) {
	n.Props[key] = value
}

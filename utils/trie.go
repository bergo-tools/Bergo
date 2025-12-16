package utils

type node struct {
	value    interface{}
	children map[rune]*node
}

type Trie struct {
	root *node
}

func NewTrie() *Trie {
	return &Trie{
		root: &node{
			children: make(map[rune]*node),
		},
	}
}

func (t *Trie) Put(key string, value interface{}) {
	current := t.root
	for _, ch := range key {
		if current.children[ch] == nil {
			current.children[ch] = &node{
				children: make(map[rune]*node),
			}
		}
		current = current.children[ch]
	}
	current.value = value
}

func (t *Trie) Get(key string) interface{} {
	current := t.root
	for _, ch := range key {
		if current.children[ch] == nil {
			return nil
		}
		current = current.children[ch]
	}
	return current.value
}

func (t *Trie) Delete(key string) {
	var deleteNodes []*node
	current := t.root

	for _, ch := range key {
		if current.children[ch] == nil {
			return
		}
		deleteNodes = append(deleteNodes, current)
		current = current.children[ch]
	}
	current.value = nil

	for i := len(deleteNodes) - 1; i >= 0; i-- {
		parent := deleteNodes[i]
		ch := rune(key[i])
		if len(current.children) == 0 && current.value == nil {
			delete(parent.children, ch)
		}
		current = parent
	}
}

func (t *Trie) Walk(walkFunc func(key string, value interface{})) {
	t.walkHelper(t.root, "", walkFunc)
}

func (t *Trie) walkHelper(n *node, prefix string, walkFunc func(key string, value interface{})) {
	if n.value != nil {
		walkFunc(prefix, n.value)
	}
	for ch, child := range n.children {
		t.walkHelper(child, prefix+string(ch), walkFunc)
	}
}

func (t *Trie) WalkPath(prefix string, walkFunc func(key string, value interface{})) {
	current := t.root
	for _, ch := range prefix {
		if current.children[ch] == nil {
			return
		}
		current = current.children[ch]
	}
	t.walkHelper(current, prefix, walkFunc)
}

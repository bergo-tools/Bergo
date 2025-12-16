package test

import (
	"testing"

	"bergo/utils"
)

func TestTrieOperations(t *testing.T) {
	trie := utils.NewTrie()

	// 测试Put和Get
	trie.Put("apple", 1)
	trie.Put("appt", 2)
	trie.Put("appc", 3)
	trie.Put("test", 4)
	if val := trie.Get("apple"); val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	// 测试不存在的键
	if val := trie.Get("app"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	// 测试Walk
	count := 0
	trie.Walk(func(key string, value interface{}) {
		count++
	})
	if count != 4 {
		t.Errorf("Expected 4 item, got %d", count)
	}

	prefixCount := 0
	trie.WalkPath("app", func(key string, value interface{}) {
		prefixCount++
	})
	if prefixCount != 3 {
		t.Errorf("Expected 3 items under prefix, got %d", prefixCount)
	}

	// 测试Delete
	trie.Delete("apple")
	if val := trie.Get("apple"); val != nil {
		t.Errorf("Expected nil after deletion, got %v", val)
	}
}

func TestFIlepath(t *testing.T) {
	np := utils.NewNowPath()
	prefix := ""
	np.Update(prefix)
	for _, file := range np.MatchFiles(prefix) {
		t.Log(file)
	}
}

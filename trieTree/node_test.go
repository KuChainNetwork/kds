package trieTree

import (
	"testing"
)

func TestNode_Search(t *testing.T) {
	trieTree := NewNode().Add("/api/v1/user", func() {
		t.Log("/api/v1/user")
	})
	node := trieTree.Search("/b")
	if nil != node {
		t.Error("Search /b")
		return
	}
	if node = trieTree.Search("/api/v1"); nil == node {
		t.Error("Search /api/v1")
		return
	}
	if node = trieTree.Search("/api/v1/user"); nil == node {
		t.Error("Search /api/v1/user")
		return
	}
	trieTree.Search("/api/v1/user")[0].Data.(func())()
}

func TestNode_StartWith(t *testing.T) {
	list := NewNode().Add("/api/v1/a", nil).
		Add("/api/v1/b", nil).
		Add("/api/v1/c", nil).
		StartWith("/api/v1", 3)
	if nil == list || 0 >= len(list) {
		t.Error("StartWith")
		return
	}
	t.Log(list)
}

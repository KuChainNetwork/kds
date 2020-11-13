package trieTree

// Node 前缀树节点
type Node struct {
	Char string      `json:"char"`
	Data interface{} `json:"data"`
	Next []*Node     `json:"next"`
}

// NewNode 工厂方法
func NewNode() *Node {
	return &Node{Next: make([]*Node, 0, 0)}
}

// Add
func (object *Node) Add(word string, data interface{}) *Node {
	if 0 >= len(word) {
		return object
	}
	next := object.Next
	ptrNext := &object.Next
	for i := 0; i < len(word); i++ {
		ch := string(word[i])
		found := false
		for _, node := range next {
			if ch == node.Char {
				found = true
				next = node.Next
				ptrNext = &node.Next
				if i >= len(word)-1 && 0 < len(*ptrNext) {
					(*ptrNext)[len(*ptrNext)-1].Data = data
				}
				break
			}
		}
		if !found {
			*ptrNext = append(*ptrNext, &Node{Char: ch})
			if i >= len(word)-1 {
				(*ptrNext)[len(*ptrNext)-1].Data = data
			}
			next = (*ptrNext)[len(*ptrNext)-1].Next
			ptrNext = &(*ptrNext)[len(*ptrNext)-1].Next
		}
	}
	return object
}

// AddAll
func (object *Node) AddAll(words []string, data interface{}) *Node {
	for _, word := range words {
		object.Add(word, data)
	}
	return object
}

// search
func (object *Node) search(word string, index int, list *[]*Node) {
	if nil != object.Next && 0 < len(object.Next) {
		for i := 0; i < len(object.Next); i++ {
			if object.Next[i].Char[0] == word[index] {
				if index < len(word)-1 {
					object.Next[i].search(word, index+1, list)
				} else {
					*list = append(*list, object.Next[i])
				}
			}
		}
	}
}

// Search
func (object *Node) Search(word string) (list []*Node) {
	object.search(word, 0, &list)
	return
}

// startWith
func (object *Node) startWith(prefix string, max int, filter func(word string, list []string) bool, list *[]string) {
	if nil != object.Next && 0 < len(object.Next) {
		for i := 0; i < len(object.Next); i++ {
			object.Next[i].startWith(prefix+object.Char, max, filter, list)
		}
	}
	if len(*list) >= max {
		return
	}
	if filter(prefix+object.Char, *list) {
		*list = append(*list, prefix+object.Char)
	}
}

// StartWith
func (object *Node) StartWith(word string, max int, filter func(word string, list []string) bool) (list []string) {
	nodeList := object.Search(word)
	if nil == nodeList || 0 >= len(nodeList) {
		return
	}
	for _, node := range nodeList {
		if filter(word, list) {
			list = append(list, word)
		}
		if nil == node.Next || 0 >= len(node.Next) {
			if filter(word, list) {
				list = append(list, word)
			}
			continue
		}
		for _, curr := range node.Next {
			curr.startWith(word, max, filter, &list)
			if len(list) >= max {
				return
			}
		}
	}
	return
}

// Traverse
func (object *Node) Traverse(cb func(node *Node) (stop bool)) {
	for _, node := range object.Next {
		if nil != node {
			if cb(node) {
				return
			}
			node.Traverse(cb)
		}
	}
}

package inotify

import (
	"path"
	"strings"
)

type node struct {
	child map[string]*node
	depth int
	end   bool
	path  string
	docId string
}

func (n *node) deletePath(dirPath string) []string {
	files := make([]string, 0)
	dirPath = strings.TrimSuffix(dirPath, "/")
	parent := n.getNodeParent(dirPath)
	if parent == nil {
		return files
	}
	if parent.child == nil {
		return files
	}
	dir, ok := parent.child[path.Base(dirPath)]
	if !ok {
		return files
	}
	delete(parent.child, path.Base(dirPath))
	return dir.getLeafs()
}

func (n *node) getLeafs() []string {
	leafs := make([]string, 0)
	if n.end {
		leafs = append(leafs, n.path)
	}
	if n.child == nil {
		return leafs
	}
	for _, clildNode := range n.child {
		leafs = append(leafs, clildNode.getLeafs()...)
	}
	return leafs
}

func (n *node) getNodeParent(filePath string) *node {
	wordData := strings.Split(filePath, "/")
	wordLength := len(wordData)
	if wordLength < 2 {
		return nil
	}
	insNode := n
	for i := 1; i < wordLength; i++ {
		if insNode.child == nil {
			return nil
		}
		d := wordData[i]
		child, ok := insNode.child[d]
		if !ok {
			return nil
		}
		if i == wordLength-1 {
			return insNode
		}
		insNode = child
	}
	return nil
}

func (n *node) addFile(filePath string) {
	wordData := strings.Split(filePath, "/")
	insNode := n
	wordLength := len(wordData)
	insPath := ""
	if wordLength < 2 {
		return
	}
	for i := 1; i < wordLength; i++ {
		d := wordData[i]
		insPath = insPath + "/" + d
		if insNode.child == nil {
			insNode.child = make(map[string]*node)
			newChild := new(node)
			newChild.path = insPath
			newChild.depth = insNode.depth + 1
			if i == wordLength-1 {
				newChild.end = true
			}
			insNode.child[d] = newChild
			insNode = newChild
			continue
		}
		if child, ok := insNode.child[d]; ok {
			insNode = child
			continue
		}
		newChild := new(node)
		newChild.path = insPath
		newChild.depth = insNode.depth + 1
		insNode.child[d] = newChild
		insNode = newChild
	}
	insNode.end = true
}

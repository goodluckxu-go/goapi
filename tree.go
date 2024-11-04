package goapi

import (
	"fmt"
	"regexp"
	"strings"
)

type node struct {
	// It is a matching prefix
	prefix string

	// Is it an exact matching prefix
	isExact bool

	// Do you want to abort the matching
	isStop bool

	// Is it a static file
	isStatic bool

	// A path that requires precise matching
	fixedPaths []string

	// Match the parameter names in the path
	pathParams []string

	// It is a processed executable routing function
	router *appRouter

	// It is a child node that has successfully matched the prefix
	children []*node
}

func (n *node) getPrefix(path string) (prefix, other string) {
	if path[0] != '/' {
		return
	}
	idx := strings.Index(path[1:], "/")
	if idx == -1 {
		prefix = path
		return
	}
	prefix = path[:idx+1]
	other = path[idx+1:]
	return
}

func (n *node) addRouter(path string, router *appRouter) (err error) {
	if path == "" {
		return
	}
	prefix, other := n.getPrefix(path)
	tree := &node{
		prefix: prefix,
	}
	if other == "" {
		tree.isStop = true
		tree.router = router
	}
	if router.isPrefix {
		// It is a static resource file path
		tree.isExact = true
		tree.isStatic = true
	} else {
		left := strings.Index(prefix, "{")
		right := strings.Index(prefix, "}")
		if left == -1 && right == -1 {
			tree.isExact = true
		} else {
			for {
				if left == -1 && right == -1 {
					break
				}
				if (left == -1 && right != -1) || (left != -1 && right == -1) || left > right {
					// If the parentheses containing variable parameters are not a pair, it indicates
					// that the path definition is incorrect
					err = fmt.Errorf("path format error")
					return
				}
				fixed := prefix[:left]
				param := prefix[left+1 : right]
				if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(param) {
					// Variable parameters can only contain uppercase and lowercase numbers and underscores
					err = fmt.Errorf("path format error")
					return
				}
				prefix = prefix[right+1:]
				tree.fixedPaths = append(tree.fixedPaths, fixed)
				tree.pathParams = append(tree.pathParams, param)
				left = strings.Index(prefix, "{")
				right = strings.Index(prefix, "}")
			}
			if prefix != "" {
				tree.fixedPaths = append(tree.fixedPaths, prefix)
			}
		}
	}
	idx := n.findChildren(tree)
	if idx != -1 {
		// If the value of the prefix already exists, merge it
		tmpTree := n.children[idx]
		tmpTree.isStop = tree.isStop || tmpTree.isStop
		tmpTree.isStatic = tree.isStatic || tmpTree.isStatic
		if tmpTree.router == nil {
			tmpTree.router = tree.router
		}
		tree = tmpTree
	}
	if other != "" {
		if err = tree.addRouter(other, router); err != nil {
			return
		}
	}
	if idx != -1 {
		n.children[idx] = tree
	} else {
		n.children = append(n.children, tree)
	}
	return
}

func (n *node) findChildren(tree *node) int {
	for k, v := range n.children {
		if v.prefix == tree.prefix {
			return k
		}
	}
	return -1
}

func (n *node) findRouter(urlPath string) (router *appRouter, paths map[string]string, exists bool) {
	if urlPath == "" {
		return
	}
	oldPrefix, other := n.getPrefix(urlPath)
out:
	for _, v := range n.children {
		prefix := oldPrefix
		paths = map[string]string{}
		if v.isExact && prefix == v.prefix {
			// Accurately match prefixes
			if v.isStatic && v.isStop {
				// Static resource file judgment
				router = v.router
				exists = true
				return
			}
			if other != "" {
				// Unfinished recursive judgment
				if childRouter, childPaths, childExists := v.findRouter(other); childExists {
					for key, val := range childPaths {
						paths[key] = val
					}
					router = childRouter
					exists = true
					return
				} else {
					continue
				}
			}
			if v.isStop {
				router = v.router
				exists = true
				return
			}
		}
		if !v.isExact {
			// When performing fuzzy matching
			fixLeft := 0    // Index of the current prefix fixed string
			paramLeft := -1 // Index of the current prefix variable parameter
			// There will definitely be a fixed string stored first
			// example: /user_{id}_{name}_info
			// First, search for the first fixed value. Once found, add 1 to both indexes,
			// and then search for the next fixed value. The value above the next fixed value is the
			// value of the previous variable parameter
			for fixLeft < len(v.fixedPaths) && paramLeft < len(v.pathParams) {
				idx := strings.Index(prefix, v.fixedPaths[fixLeft])
				if idx == -1 {
					paths = nil
					continue out
				}
				if paramLeft != -1 {
					paths[v.pathParams[paramLeft]] = prefix[:idx]
				}
				prefix = prefix[idx+len(v.fixedPaths[fixLeft]):]
				fixLeft++
				paramLeft++
			}
			if fixLeft < len(v.fixedPaths) {
				// When the last parameter is a fixed string
				idx := strings.Index(prefix, v.fixedPaths[fixLeft])
				if idx == -1 {
					paths = nil
					continue out
				}
				paths[v.pathParams[paramLeft]] = prefix[:idx]
				fixLeft++
				paramLeft++
			} else if paramLeft < len(v.pathParams) && paramLeft != -1 {
				// When the last parameter is a variable parameter
				paths[v.pathParams[paramLeft]] = prefix
				paramLeft++
			}
			if fixLeft == len(v.fixedPaths) && paramLeft == len(v.pathParams) {
				// When both fixed and variable parameters have been determined, the matching is successful
				if other != "" {
					// Unfinished recursive judgment
					if childRouter, childPaths, childExists := v.findRouter(other); childExists {
						for key, val := range childPaths {
							paths[key] = val
						}
						router = childRouter
						exists = true
						return
					} else {
						continue
					}
				}
				if v.isStop {
					router = v.router
					exists = true
					return
				}
			}
		}
	}
	return
}

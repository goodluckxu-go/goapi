package goapi

import (
	"errors"
	"strings"
)

type HandleFunc func(ctx *Context)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// Get returns the value of the first Param which key matches the given name and a boolean true.
// If no matching Param is found, an empty string is returned and a boolean false .
func (ps Params) Get(name string) (string, bool) {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return "", false
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) (va string) {
	va, _ = ps.Get(name)
	return
}

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

type nodeType uint8

const (
	static nodeType = iota
	param
	catchAll
)

type trees []*node

type node struct {
	// matching string
	path string

	// the complete path
	fullPath string

	// current matching node type
	nType nodeType

	// child nodes are sorted by the number of handles registered by their child nodes, with wildcards ranked last
	children []*node

	// handler function
	handler HandleFunc

	// the priority of nodes, used for priority judgment during sorting and matching
	priority uint32

	// index of the first letter of a child string
	indices []byte

	// is there a wildcard child present
	isWildcard bool

	// Wildcard param
	params []string
}

func (n *node) min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func (n *node) longestCommonPrefix(a, b string) int {
	i := 0
	imax := n.min(len(a), len(b))
	for i < imax && a[i] == b[i] {
		i++
	}
	return i
}

// Search for a wildcard segment and check the name for invalid characters.
// Returns true as valid and -1 as index, if no wildcard was found.
func (n *node) findWildcard(path string) (wildcard string, i int, valid bool) {
	i = -1
	valid = true
	for j, c := range []byte(path) {
		switch c {
		case '{':
			i = j
		case '}':
			if i == -1 {
				return path[:j+1], i, false
			}
			return path[i : j+1], i, valid
		case '/':
			if i != -1 {
				valid = false
			}
		}
	}
	return "", -1, true
}

func (n *node) parseWildcard(wildcard string) (rs string, nType nodeType) {
	wildcard = wildcard[1 : len(wildcard)-1]
	rs = strings.TrimSuffix(wildcard, ":*")
	nType = param
	if rs != wildcard {
		nType = catchAll
	}
	return
}

// addChild will add a child node, keeping wildcardChild at the end
func (n *node) addChild(child *node) {
	if len(n.children) > 0 && n.children[len(n.children)-1].nType != static {
		n.children = append(n.children[:len(n.children)-1], child, n.children[len(n.children)-1])
	} else {
		n.children = append(n.children, child)
	}
}

// addRoute adds a node with the given handle to the path.
func (n *node) addRoute(path string, handler HandleFunc) (err error) {
	fullPath := path
	var params []string
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		err = n.insertChild(path, fullPath, params, handler)
		return
	}

walk:
	for {
		// Find the longest common prefix.
		// This also implies that the common prefix contains no wildcards
		// since the existing key can't contain those chars.
		i := n.longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:       n.path[i:],
				isWildcard: n.isWildcard,
				nType:      static,
				indices:    n.indices,
				children:   n.children,
				handler:    n.handler,
				priority:   n.priority - 1,
				fullPath:   n.fullPath,
				params:     n.params,
			}

			n.children = []*node{&child}
			n.indices = []byte{n.path[i]}
			n.path = path[:i]
			n.handler = nil
			n.isWildcard = false
			n.params = nil
			n.fullPath = ""
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// Check if a child with the next path byte exists
			for i := 0; i < len(n.indices); i++ {
				if c == n.indices[i] {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if c != '{' {
				n.indices = append(n.indices, c)
				child := &node{
					priority: 1,
				}
				n.addChild(child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			} else if n.isWildcard {
				// inserting a wildcard node, need to check if it conflicts with the existing wildcard
				wildcard := path[:strings.IndexByte(path, '}')+1]
				paramStr, nType := n.parseWildcard(wildcard)
				params = append(params, paramStr)
				n = n.children[len(n.children)-1]
				n.priority++
				path = path[len(wildcard):]

				if n.nType == catchAll || nType == catchAll {
					err = errors.New("the 'catch-all' wildcard can only exist once, has '" + wildcard +
						"' in path '" + fullPath + "'")
					return
				}

				continue walk
			}

			err = n.insertChild(path, fullPath, params, handler)
			return
		}

		// Otherwise add handle to current node
		if n.handler != nil {
			err = errors.New("handlers are already registered for path '" + fullPath + "'")
			return
		}
		n.handler = handler
		n.fullPath = fullPath
		n.params = params
		return
	}
}

func (n *node) insertChild(path string, fullPath string, params []string, handler HandleFunc) (err error) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := n.findWildcard(path)

		// Wildcards must be between ‘{’ and ‘}’ and the ‘/’ character must not exist
		if !valid {
			err = errors.New("wildcards must be between ‘{’ and ‘}’ and the ‘/’ character must not exist, has: '" +
				wildcard + "' in path '" + fullPath + "'")
			return
		}

		if i < 0 { // No wildcard found
			break
		}

		// check if the wildcard has a name
		if len(wildcard) < 3 {
			err = errors.New("wildcards must be named with a non-empty name, has '" + wildcard + "' in path '" +
				fullPath + "'")
			return
		}

		if i > 0 {
			if path[i-1] != '/' {
				err = errors.New("no / before wildcards '" + string(path[i-1]) + "' in path '" + fullPath + "'")
				return
			}
			n.path = path[:i]
			path = path[i:]
		}

		paramStr, nType := n.parseWildcard(wildcard)
		params = append(params, paramStr)

		// if the path doesn't end with the wildcard, then there
		// will be another subpath starting with '/'
		if len(wildcard) < len(path) {
			if nType == catchAll {
				err = errors.New("'catch-all' wildcard are only allowed at the end of the path in path '" + fullPath + "'")
				return
			}
			if path[len(wildcard)] != '/' {
				err = errors.New("no / after wildcards '" + string(path[len(wildcard)]) + "' in path '" + fullPath + "'")
				return
			}
			// param
			path = path[len(wildcard):]
			child := &node{
				nType:    nType,
				priority: 1,
				indices:  []byte{path[0]},
			}
			n.addChild(child)
			n.isWildcard = true
			n = child
			// surplus path
			child = &node{
				nType:    static,
				priority: 1,
			}
			n.addChild(child)
			n = child
			continue
		}

		// The last one is the wildcard node
		child := &node{
			nType:    nType,
			handler:  handler,
			fullPath: fullPath,
			params:   params,
			priority: 1,
		}
		n.addChild(child)
		n.isWildcard = true
		return
	}
	// If no wildcard was found, simply insert the path and handle
	n.path = path
	n.handler = handler
	n.fullPath = fullPath
	n.params = params
	return
}

// Increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Build new index char string
	if newPos != pos {
		n.indices = n.addBytes(n.indices[:newPos], // Unchanged prefix, might be empty
			n.indices[pos:pos+1], // The index char we move
			n.indices[newPos:pos], n.indices[pos+1:]) // Rest without char at 'pos'
	}

	return newPos
}

func (n *node) addBytes(vals ...[]byte) (rs []byte) {
	for _, val := range vals {
		rs = append(rs, val...)
	}
	return
}

// nodeValue holds return values of (*Node).getValue method
type nodeValue struct {
	handler  HandleFunc
	params   Params
	fullPath string
	tsr      bool
}

type skippedNode struct {
	path   string
	node   *node
	params []string
}

func (n *node) getValue(path string) (value nodeValue) {
	var skippedNodes []skippedNode
	var params []string
walk:
	for {
		// When matching ‘cache-all’, it will be returned directly
		if n.nType == catchAll {
			params = append(params, path)
			n.returnValue(params, &value)
			return
		}

		isMatch := false
		if n.nType == param {
			// match param
			i := 0
			for ; i < len(path) && path[i] != '/'; i++ {
			}
			params = append(params, path[:i])
			path = path[i:]
			isMatch = true
		} else if n.nType == static && len(path) >= len(n.path) && path[:len(n.path)] == n.path {
			// match static
			path = path[len(n.path):]
			isMatch = true
		}
		if !value.tsr {
			if isMatch {
				if path == "" {
					// The matching has been completed. Check if the subset matches
					for i, c := range n.indices {
						if c == '/' {
							value.tsr = n.children[i].path == "/" && n.children[i].handler != nil
							break
						}
					}
					value.tsr = value.tsr || (n.isWildcard && n.children[len(n.children)-1].handler != nil)
				} else if path == "/" {
					// If there are remaining ‘/’, it is recommended to redirect to a path without ‘/’
					value.tsr = n.handler != nil
				}
			} else if path != "" {
				// There will only be static addresses that cannot be matched. Check if the tsr address matches
				tsrPath := path
				if tsrPath[len(tsrPath)-1] == '/' {
					tsrPath = tsrPath[:len(tsrPath)-1]
				} else {
					tsrPath += "/"
				}
				value.tsr = tsrPath == n.path && n.handler != nil
			}
		}

		if isMatch && path != "" {
			// After a successful match, check if there are static child nodes.
			// If not found, add the wildcard child nodes to skippedNodes for easy rollback in case of failure
			for i, c := range n.indices {
				if c == path[0] {
					if n.isWildcard {
						skippedNodes = append(skippedNodes, skippedNode{
							path:   path,
							node:   n.children[len(n.children)-1],
							params: params,
						})
					}
					n = n.children[i]
					continue walk
				}
			}

			// If the static child node fails to match, the wildcard child node will be matched
			if n.isWildcard {
				n = n.children[len(n.children)-1]
				continue walk
			}
		}

		// When the matching is completed, return the matching value
		if path == "" {
			n.returnValue(params, &value)
			return
		}
		// Handle the matching of the rollback
		// If the rollback fails, return
		if len(skippedNodes) == 0 {
			return
		}
		skipped := skippedNodes[0]
		skippedNodes = skippedNodes[1:]
		n = skipped.node
		path = skipped.path
		params = skipped.params
	}
}

func (n *node) returnValue(params []string, valuePtr *nodeValue) {
	// The number of matching wildcards is incorrect
	if len(n.params) != len(params) {
		return
	}
	valuePtr.handler = n.handler
	valuePtr.fullPath = n.fullPath
	valuePtr.tsr = valuePtr.tsr && n.nType != static
	for i, key := range n.params {
		valuePtr.params = append(valuePtr.params, Param{
			Key:   key,
			Value: params[i],
		})
	}
	return
}

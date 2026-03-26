package goapi

import (
	"testing"
)

func fakeHandler(ctx *Context) {}

func TestParams_Get(t *testing.T) {
	ps := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "alice"},
	}

	val, ok := ps.Get("id")
	if !ok || val != "123" {
		t.Errorf("expected (\"123\", true), got (%q, %v)", val, ok)
	}

	val, ok = ps.Get("name")
	if !ok || val != "alice" {
		t.Errorf("expected (\"alice\", true), got (%q, %v)", val, ok)
	}

	val, ok = ps.Get("missing")
	if ok || val != "" {
		t.Errorf("expected (\"\", false), got (%q, %v)", val, ok)
	}
}

func TestParams_ByName(t *testing.T) {
	ps := Params{
		{Key: "id", Value: "42"},
	}

	if v := ps.ByName("id"); v != "42" {
		t.Errorf("expected \"42\", got %q", v)
	}
	if v := ps.ByName("unknown"); v != "" {
		t.Errorf("expected \"\", got %q", v)
	}
}

func TestMethodTrees_Get(t *testing.T) {
	root := &node{}
	trees := methodTrees{
		{method: "GET", root: root},
		{method: "POST", root: &node{}},
	}

	if got := trees.get("GET"); got != root {
		t.Error("expected to find GET tree root")
	}
	if got := trees.get("DELETE"); got != nil {
		t.Error("expected nil for non-existent method")
	}
}

func TestNode_Min(t *testing.T) {
	n := &node{}
	if n.min(3, 5) != 3 {
		t.Errorf("expected 3")
	}
	if n.min(7, 2) != 2 {
		t.Errorf("expected 2")
	}
	if n.min(4, 4) != 4 {
		t.Errorf("expected 4")
	}
}

func TestNode_LongestCommonPrefix(t *testing.T) {
	n := &node{}
	tests := []struct {
		a, b string
		want int
	}{
		{"abc", "abd", 2},
		{"abc", "abc", 3},
		{"abc", "xyz", 0},
		{"", "abc", 0},
		{"abc", "", 0},
		{"/users/list", "/users/detail", 7},
	}
	for _, tt := range tests {
		if got := n.longestCommonPrefix(tt.a, tt.b); got != tt.want {
			t.Errorf("longestCommonPrefix(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestNode_FindWildcard(t *testing.T) {
	n := &node{}

	tests := []struct {
		path      string
		wildcard  string
		index     int
		valid     bool
	}{
		{"/static/path", "", -1, true},
		{"/{id}", "{id}", 1, true},
		{"/{id}/detail", "{id}", 1, true},
		{"/{id:*}", "{id:*}", 1, true},
		// '}' without '{' is invalid
		{"/bad}", "/bad}", -1, false},
		// '/' inside wildcard is invalid
		{"/{a/b}", "{a/b}", 1, false},
	}
	for _, tt := range tests {
		wildcard, i, valid := n.findWildcard(tt.path)
		if wildcard != tt.wildcard || i != tt.index || valid != tt.valid {
			t.Errorf("findWildcard(%q) = (%q, %d, %v), want (%q, %d, %v)",
				tt.path, wildcard, i, valid, tt.wildcard, tt.index, tt.valid)
		}
	}
}

func TestNode_ParseWildcard(t *testing.T) {
	n := &node{}

	rs, nType := n.parseWildcard("{id}")
	if rs != "id" || nType != param {
		t.Errorf("expected (\"id\", param), got (%q, %v)", rs, nType)
	}

	rs, nType = n.parseWildcard("{path:*}")
	if rs != "path" || nType != catchAll {
		t.Errorf("expected (\"path\", catchAll), got (%q, %v)", rs, nType)
	}
}

func TestNode_AddBytes(t *testing.T) {
	n := &node{}
	result := n.addBytes([]byte("ab"), []byte("cd"), []byte("ef"))
	if string(result) != "abcdef" {
		t.Errorf("expected \"abcdef\", got %q", string(result))
	}

	result = n.addBytes()
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestNode_AddRoute_StaticPaths(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/hello", fakeHandler); err != nil {
		t.Fatalf("unexpected error adding /hello: %v", err)
	}
	if err := n.addRoute("/help", fakeHandler); err != nil {
		t.Fatalf("unexpected error adding /help: %v", err)
	}
	if err := n.addRoute("/world", fakeHandler); err != nil {
		t.Fatalf("unexpected error adding /world: %v", err)
	}

	val := n.getValue("/hello")
	if val.handler == nil || val.fullPath != "/hello" {
		t.Errorf("expected handler for /hello, got fullPath=%q handler=%v", val.fullPath, val.handler)
	}

	val = n.getValue("/help")
	if val.handler == nil || val.fullPath != "/help" {
		t.Errorf("expected handler for /help, got fullPath=%q handler=%v", val.fullPath, val.handler)
	}

	val = n.getValue("/world")
	if val.handler == nil || val.fullPath != "/world" {
		t.Errorf("expected handler for /world, got fullPath=%q handler=%v", val.fullPath, val.handler)
	}

	val = n.getValue("/notfound")
	if val.handler != nil {
		t.Errorf("expected nil handler for /notfound")
	}
}

func TestNode_AddRoute_DuplicatePath(t *testing.T) {
	n := &node{}
	if err := n.addRoute("/dup", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.addRoute("/dup", fakeHandler); err == nil {
		t.Fatal("expected error for duplicate route, got nil")
	}
}

func TestNode_AddRoute_ParamWildcard(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/users/{id}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.addRoute("/users/{id}/profile", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/users/42")
	if val.handler == nil {
		t.Fatal("expected handler for /users/42")
	}
	if val.fullPath != "/users/{id}" {
		t.Errorf("expected fullPath /users/{id}, got %q", val.fullPath)
	}
	if v := val.params.ByName("id"); v != "42" {
		t.Errorf("expected param id=42, got %q", v)
	}

	val = n.getValue("/users/100/profile")
	if val.handler == nil {
		t.Fatal("expected handler for /users/100/profile")
	}
	if val.fullPath != "/users/{id}/profile" {
		t.Errorf("expected fullPath /users/{id}/profile, got %q", val.fullPath)
	}
	if v := val.params.ByName("id"); v != "100" {
		t.Errorf("expected param id=100, got %q", v)
	}
}

func TestNode_AddRoute_CatchAll(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/files/{path:*}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/files/a/b/c")
	if val.handler == nil {
		t.Fatal("expected handler for /files/a/b/c")
	}
	if v := val.params.ByName("path"); v != "a/b/c" {
		t.Errorf("expected param path=a/b/c, got %q", v)
	}
}

func TestNode_AddRoute_CatchAllNotAtEnd(t *testing.T) {
	n := &node{}
	err := n.addRoute("/files/{path:*}/extra", fakeHandler)
	if err == nil {
		t.Fatal("expected error for catch-all not at end of path")
	}
}

func TestNode_AddRoute_InvalidWildcard(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"empty wildcard name", "/users/{}"},
		{"no slash before wildcard", "/users{id}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{}
			if err := n.addRoute(tt.path, fakeHandler); err == nil {
				t.Errorf("expected error for path %q, got nil", tt.path)
			}
		})
	}
}

func TestNode_AddRoute_MultipleParams(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/api/{version}/{resource}/{id}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/api/v1/users/42")
	if val.handler == nil {
		t.Fatal("expected handler for /api/v1/users/42")
	}
	if v := val.params.ByName("version"); v != "v1" {
		t.Errorf("expected param version=v1, got %q", v)
	}
	if v := val.params.ByName("resource"); v != "users" {
		t.Errorf("expected param resource=users, got %q", v)
	}
	if v := val.params.ByName("id"); v != "42" {
		t.Errorf("expected param id=42, got %q", v)
	}
}

func TestNode_AddRoute_StaticAndParamCoexist(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/users/list", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.addRoute("/users/{id}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/users/list")
	if val.handler == nil || val.fullPath != "/users/list" {
		t.Errorf("expected /users/list handler, got fullPath=%q", val.fullPath)
	}

	val = n.getValue("/users/99")
	if val.handler == nil || val.fullPath != "/users/{id}" {
		t.Errorf("expected /users/{id} handler, got fullPath=%q", val.fullPath)
	}
}

func TestNode_GetValue_TSR(t *testing.T) {
	n := &node{}

	if err := n.addRoute("/foo/", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/foo")
	if !val.tsr {
		t.Error("expected tsr=true for /foo when /foo/ is registered")
	}

	n2 := &node{}
	if err := n2.addRoute("/bar", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val = n2.getValue("/bar/")
	if !val.tsr {
		t.Error("expected tsr=true for /bar/ when /bar is registered")
	}
}

func TestNode_GetValue_NotFound(t *testing.T) {
	n := &node{}
	if err := n.addRoute("/exists", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/nope")
	if val.handler != nil {
		t.Error("expected nil handler for unregistered path")
	}
}

func TestNode_GetValue_EmptyTree(t *testing.T) {
	n := &node{}
	val := n.getValue("/anything")
	if val.handler != nil {
		t.Error("expected nil handler on empty tree")
	}
}

func TestNode_AddRoute_ComplexTree(t *testing.T) {
	n := &node{}

	routes := []string{
		"/",
		"/api/v1/users",
		"/api/v1/users/{id}",
		"/api/v1/users/{id}/posts",
		"/api/v1/posts",
		"/api/v2/users",
		"/static/{filepath:*}",
	}

	for _, r := range routes {
		if err := n.addRoute(r, fakeHandler); err != nil {
			t.Fatalf("unexpected error adding %q: %v", r, err)
		}
	}

	tests := []struct {
		path     string
		found    bool
		fullPath string
	}{
		{"/", true, "/"},
		{"/api/v1/users", true, "/api/v1/users"},
		{"/api/v1/users/10", true, "/api/v1/users/{id}"},
		{"/api/v1/users/10/posts", true, "/api/v1/users/{id}/posts"},
		{"/api/v1/posts", true, "/api/v1/posts"},
		{"/api/v2/users", true, "/api/v2/users"},
		{"/static/css/style.css", true, "/static/{filepath:*}"},
		{"/api/v3/users", false, ""},
		{"/missing", false, ""},
	}

	for _, tt := range tests {
		val := n.getValue(tt.path)
		if tt.found && val.handler == nil {
			t.Errorf("expected handler for %q, got nil", tt.path)
		} else if !tt.found && val.handler != nil {
			t.Errorf("expected nil handler for %q", tt.path)
		}
		if tt.found && val.fullPath != tt.fullPath {
			t.Errorf("path %q: expected fullPath=%q, got %q", tt.path, tt.fullPath, val.fullPath)
		}
	}
}

func TestNode_AddRoute_DuplicateCatchAll(t *testing.T) {
	n := &node{}
	if err := n.addRoute("/files/{path:*}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := n.addRoute("/files/{other:*}", fakeHandler)
	if err == nil {
		t.Fatal("expected error for duplicate catch-all wildcard")
	}
}

func TestNode_AddChild_KeepsWildcardLast(t *testing.T) {
	root := &node{}
	wildcardChild := &node{nType: param}
	staticChild1 := &node{nType: static, path: "/a"}
	staticChild2 := &node{nType: static, path: "/b"}

	root.children = append(root.children, wildcardChild)

	root.addChild(staticChild1)
	last := root.children[len(root.children)-1]
	if last.nType != param {
		t.Error("expected wildcard child to remain at the end")
	}

	root.addChild(staticChild2)
	last = root.children[len(root.children)-1]
	if last.nType != param {
		t.Error("expected wildcard child to remain at the end after second insert")
	}
}

func TestNode_IncrementChildPrio(t *testing.T) {
	root := &node{
		indices: []byte{'a', 'b', 'c'},
		children: []*node{
			{path: "a", priority: 3},
			{path: "b", priority: 1},
			{path: "c", priority: 2},
		},
	}

	// Incrementing 'c' (index 2) priority to 3 should reorder it
	newPos := root.incrementChildPrio(2)
	if newPos > 2 {
		t.Errorf("expected newPos <= 2, got %d", newPos)
	}
	// After increment, child at newPos should have priority 3
	if root.children[newPos].priority != 3 {
		t.Errorf("expected priority 3 at newPos, got %d", root.children[newPos].priority)
	}
}

func TestNode_GetValue_SkippedNodes_Backtrack(t *testing.T) {
	n := &node{}

	// Register a param route and a deeper static route under a param
	if err := n.addRoute("/a/{id}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.addRoute("/a/{id}/details", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.addRoute("/a/special", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "special" should match the static route, not the param route
	val := n.getValue("/a/special")
	if val.handler == nil {
		t.Fatal("expected handler for /a/special")
	}
	if val.fullPath != "/a/special" {
		t.Errorf("expected fullPath /a/special, got %q", val.fullPath)
	}

	val = n.getValue("/a/something")
	if val.handler == nil {
		t.Fatal("expected handler for /a/something")
	}
	if val.fullPath != "/a/{id}" {
		t.Errorf("expected fullPath /a/{id}, got %q", val.fullPath)
	}
}

func TestNode_GetValue_ParamWithTrailingSlash(t *testing.T) {
	n := &node{}
	if err := n.addRoute("/items/{id}", fakeHandler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := n.getValue("/items/5/")
	// Should suggest trailing slash redirect
	if !val.tsr {
		t.Error("expected tsr=true for /items/5/ when /items/{id} is registered")
	}
}

func TestNode_FindWildcard_NoWildcard(t *testing.T) {
	n := &node{}
	wildcard, i, valid := n.findWildcard("/plain/path/no/wildcard")
	if wildcard != "" || i != -1 || !valid {
		t.Errorf("expected (\"\", -1, true), got (%q, %d, %v)", wildcard, i, valid)
	}
}

func TestNode_AddRoute_EmptyTree(t *testing.T) {
	n := &node{}
	if err := n.addRoute("/", fakeHandler); err != nil {
		t.Fatalf("unexpected error adding root path: %v", err)
	}
	val := n.getValue("/")
	if val.handler == nil {
		t.Error("expected handler for root path")
	}
}

func TestNode_ReturnValue_ParamCountMismatch(t *testing.T) {
	n := &node{
		handler:  fakeHandler,
		fullPath: "/test/{a}/{b}",
		params:   []string{"a", "b"},
	}
	var value nodeValue
	// Only pass one param when two are expected
	n.returnValue([]string{"val1"}, &value)
	if value.handler != nil {
		t.Error("expected nil handler when param count mismatches")
	}
}

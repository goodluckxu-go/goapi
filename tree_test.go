package goapi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	// valid addRouter
	addList := []struct {
		in struct {
			path   string
			router *appRouter
		}
		out error
	}{
		{struct {
			path   string
			router *appRouter
		}{"/", &appRouter{path: "/"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user", &appRouter{path: "/user"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/list/search", &appRouter{path: "/user/list/search"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/list", &appRouter{path: "/user/list"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/edit/{id}_{name}", &appRouter{path: "/user/edit/{id}_{name}"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/edit/{id}/file", &appRouter{path: "/user/edit/{id}/file"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/edit/{id}_{name}/{uid}", &appRouter{path: "/user/edit/{id}_{name}/{uid}"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/articles", &appRouter{path: "/articles"}}, nil},
		{struct {
			path   string
			router *appRouter
		}{"/user/post/{}", &appRouter{path: "/user/post/{}"}}, fmt.Errorf("path format error")},
	}
	tr := &node{}
	for _, add := range addList {
		out := tr.addRouter(add.in.path, add.in.router)
		assert.Equal(t, add.out, out)
	}
	// valid findRouter
	findList := []struct {
		in   string
		out1 *appRouter
		out2 map[string]string
		out3 bool
	}{
		{"/", addList[0].in.router, map[string]string{}, true},
		{"/user", addList[1].in.router, map[string]string{}, true},
		{"/user/list/search", addList[2].in.router, map[string]string{}, true},
		{"/user/list", addList[3].in.router, map[string]string{}, true},
		{"/user/edit/1_2", addList[4].in.router, map[string]string{"id": "1", "name": "2"}, true},
		{"/user/edit/3", nil, map[string]string{}, false},
		{"/user/edit/3/file", addList[5].in.router, map[string]string{"id": "3"}, true},
		{"/user/edit/1_2/3", addList[6].in.router, map[string]string{"id": "1", "name": "2", "uid": "3"}, true},
		{"/articles", addList[7].in.router, map[string]string{}, true},
	}
	for _, find := range findList {
		out1, out2, out3 := tr.findRouter(find.in)
		assert.Equal(t, find.out1, out1)
		assert.Equal(t, find.out2, out2)
		assert.Equal(t, find.out3, out3)
	}
}

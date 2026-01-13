package goapi

import "github.com/goodluckxu-go/goapi/openapi"

// Router is used to set access routes and routing methods
//
//	Tag Description:
//		path: Access Routing. Multiple contents separated by ','
//		method: Access method. Multiple contents separated by ','
//		summary: A short summary of the API.
//		desc: A description of the API. CommonMark syntax MAY be used for rich text representation.
//		tags: Multiple contents separated by ','
//		deprecated: For example deprecated:"true", discard this route
type Router struct{}

type RouterTags interface {
	Tags() []*openapi.Tag
}

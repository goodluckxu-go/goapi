package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// OpenAPI is the root of an OpenAPI v3.1.0 document
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md
type OpenAPI struct {
	// REQUIRED. This string MUST be the version number of the OpenAPI Specification that the
	// OpenAPI document uses. The field SHOULD be used by tooling to interpret the OpenAPI
	// document. This is not related to the API info.version string.openapi
	OpenAPI string `json:"openapi"`

	// REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Info *Info `json:"info"`

	// The default value for the keyword within Schema Objects contained within this OAS document.
	// This MUST be in the form of a URI.$schema
	JSONSchemaDialect string `json:"jsonSchemaDialect"`

	// An array of Server Objects, which provide connectivity information to a target server.
	// If the property is not provided, or is an empty array, the default value would be a Server
	// Object with a url value of .servers/
	Servers []*Server `json:"servers"`

	// The available paths and operations for the API.
	Paths *Paths `json:"paths"`

	// The incoming webhooks that MAY be received as part of this API and that the API consumer MAY
	// choose to implement. Closely related to the feature, this section describes requests initiated
	// other than by an API call, for example by an out of band registration. The key name is a unique
	// string to refer to each webhook, while the (optionally referenced) Path Item Object describes
	// a request that may be initiated by the API provider and the expected responses. An example is
	// available.callbacks
	Webhooks map[string]*PathItem `json:"webhooks"`

	// An element to hold various schemas for the document.
	Components *Components `json:"components"`

	// A declaration of which security mechanisms can be used across the API. The list of values includes
	// alternative security requirement objects that can be used. Only one of the security requirement
	// objects need to be satisfied to authorize a request. Individual operations can override this definition.
	// To make security optional, an empty security requirement () can be included in the array.{}
	Security []*SecurityRequirement `json:"security"`

	// A list of tags used by the document with additional metadata. The order of the tags can be used to
	// reflect on their order by the parsing tools. Not all tags that are used by the Operation Object must
	// be declared. The tags that are not declared MAY be organized randomly or based on the tools' logic.
	// Each tag name in the list MUST be unique.
	Tags []*Tag `json:"tags"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (o *OpenAPI) marshalField() []marshalField {
	return []marshalField{
		{"openapi", o.OpenAPI, o.OpenAPI == ""},
		{"info", o.Info, o.Info == nil},
		{"jsonSchemaDialect", o.JSONSchemaDialect, o.JSONSchemaDialect == ""},
		{"servers", o.Servers, o.Servers == nil},
		{"paths", o.Paths, o.Paths == nil},
		{"webhooks", o.Webhooks, o.Webhooks == nil},
		{"components", o.Components, o.Components == nil},
		{"security", o.Security, o.Security == nil},
		{"tags", o.Tags, o.Tags == nil},
		{"externalDocs", o.ExternalDocs, o.ExternalDocs == nil},
	}
}

func (o *OpenAPI) MarshalJSON() ([]byte, error) {
	return marshalJson(o.marshalField(), o.Extensions)
}

func (o *OpenAPI) UnmarshalJSON(buf []byte) (err error) {
	type alias OpenAPI
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "openapi")
	delete(x.Extensions, "info")
	delete(x.Extensions, "jsonSchemaDialect")
	delete(x.Extensions, "servers")
	delete(x.Extensions, "paths")
	delete(x.Extensions, "webhooks")
	delete(x.Extensions, "components")
	delete(x.Extensions, "security")
	delete(x.Extensions, "tags")
	delete(x.Extensions, "externalDocs")
	*o = OpenAPI(x)
	return
}

func (o *OpenAPI) Validate() error {
	if !regexp.MustCompile(`^3\.1(\.\d+)*$`).MatchString(o.OpenAPI) {
		return verifyError("openapi", fmt.Errorf("must be 3.1 or 3.1.*"))
	}

	if o.Info != nil {
		if err := o.Info.Validate(); err != nil {
			return verifyError("info", err)
		}
	} else {
		return verifyError("info", fmt.Errorf("must be a non empty object"))
	}

	for k, v := range o.Servers {
		if err := v.Validate(); err != nil {
			return verifyError(fmt.Sprintf("servers[%v]", k), err)
		}
	}

	if o.Paths != nil {
		if err := o.Paths.Validate(o); err != nil {
			return verifyError("paths", err, true)
		}
	}

	for k, v := range o.Webhooks {
		if err := v.Validate(o, ""); err != nil {
			return verifyError(fmt.Sprintf("webhooks[%v]", k), err)
		}
	}

	if o.Components != nil {
		if err := o.Components.Validate(o); err != nil {
			return verifyError("components", err)
		}
	}

	for k, v := range o.Tags {
		if err := v.Validate(); err != nil {
			return verifyError(fmt.Sprintf("tags[%v]", k), err)
		}
	}

	if o.ExternalDocs != nil {
		if err := o.ExternalDocs.Validate(); err != nil {
			return verifyError("externalDocs", err)
		}
	}

	if o.Extensions != nil {
		if err := validatorExtensions(o.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`

	// A short summary of the API.
	Summary string `json:"summary"`

	// A description of the API. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// A URL to the Terms of Service for the API. This MUST be in the form of a URL.
	TermsOfService string `json:"termsOfService"`

	// The contact information for the exposed API.
	Contact *Contact `json:"contact"`

	// The license information for the exposed API.
	License *License `json:"license"`

	// REQUIRED. The version of the OpenAPI document (which is distinct from the OpenAPI Specification
	// version or the API implementation version).
	Version string `json:"version"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (i *Info) marshalField() []marshalField {
	return []marshalField{
		{"title", i.Title, i.Title == ""},
		{"summary", i.Summary, i.Summary == ""},
		{"description", i.Description, i.Description == ""},
		{"termsOfService", i.TermsOfService, i.TermsOfService == ""},
		{"contact", i.Contact, i.Contact == nil},
		{"license", i.License, i.License == nil},
		{"version", i.Version, i.Version == ""},
	}
}

func (i *Info) MarshalJSON() ([]byte, error) {
	return marshalJson(i.marshalField(), i.Extensions)
}

func (i *Info) UnmarshalJSON(buf []byte) (err error) {
	type alias Info
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "title")
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "termsOfService")
	delete(x.Extensions, "contact")
	delete(x.Extensions, "license")
	delete(x.Extensions, "version")
	*i = Info(x)
	return
}

func (i *Info) Validate() error {
	if i.Title == "" {
		return verifyError("title", fmt.Errorf("must be a non empty string"))
	}

	if i.Contact != nil {
		if err := i.Contact.Validate(); err != nil {
			return verifyError("contact", err)
		}
	}

	if i.License != nil {
		if err := i.License.Validate(); err != nil {
			return verifyError("license", err)
		}
	}

	if i.Version == "" {
		return verifyError("version", fmt.Errorf("must be a non empty string"))
	}

	if i.Extensions != nil {
		if err := validatorExtensions(i.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name"`

	// The URL pointing to the contact information. This MUST be in the form of a URL.
	URL string `json:"url"`

	// The email address of the contact person/organization. This MUST be in the form of an email address.
	Email string `json:"email"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (c *Contact) marshalField() []marshalField {
	return []marshalField{
		{"name", c.Name, c.Name == ""},
		{"url", c.URL, c.URL == ""},
		{"email", c.Email, c.Email == ""},
	}
}

func (c *Contact) MarshalJSON() ([]byte, error) {
	return marshalJson(c.marshalField(), c.Extensions)
}

func (c *Contact) UnmarshalJSON(buf []byte) (err error) {
	type alias Contact
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "name")
	delete(x.Extensions, "url")
	delete(x.Extensions, "email")
	*c = Contact(x)
	return
}

func (c *Contact) Validate() error {
	if c.Extensions != nil {
		if err := validatorExtensions(c.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type License struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name"`

	// An SPDX license expression for the API. The field is mutually exclusive of the field.identifierurl
	Identifier string `json:"identifier"`

	// A URL to the license used for the API. This MUST be in the form of a URL. The field is mutually
	// exclusive of the field.urlidentifier
	URL string `json:"url"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (l *License) marshalField() []marshalField {
	return []marshalField{
		{"name", l.Name, l.Name == ""},
		{"identifier", l.Identifier, l.Identifier == ""},
		{"url", l.URL, l.URL == ""},
	}
}

func (l *License) MarshalJSON() ([]byte, error) {
	return marshalJson(l.marshalField(), l.Extensions)
}

func (l *License) UnmarshalJSON(buf []byte) (err error) {
	type alias License
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "name")
	delete(x.Extensions, "identifier")
	delete(x.Extensions, "url")
	*l = License(x)
	return
}

func (l *License) Validate() error {
	if l.Name == "" {
		return verifyError("name", fmt.Errorf("must be a non empty string"))
	}

	if l.Identifier != "" && l.URL != "" {
		return fmt.Errorf("fields identifier and url are mutually exclusive")
	}

	if l.Extensions != nil {
		if err := validatorExtensions(l.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Server struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative,
	// to indicate that the host location is relative to the location where the OpenAPI document is
	// being served. Variable substitutions will be made when a variable is named in brackets.{}
	URL string `json:"url"`

	// An optional string describing the host designated by the URL. CommonMark syntax MAY be used
	// for rich text representation.
	Description string `json:"description"`

	// A map between a variable name and its value. The value is used for substitution in the server's URL template.
	Variables map[string]*ServerVariable `json:"variables"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (s *Server) marshalField() []marshalField {
	return []marshalField{
		{"url", s.URL, s.URL == ""},
		{"description", s.Description, s.Description == ""},
		{"variables", s.Variables, s.Variables == nil},
	}
}

func (s *Server) MarshalJSON() ([]byte, error) {
	return marshalJson(s.marshalField(), s.Extensions)
}

func (s *Server) UnmarshalJSON(buf []byte) (err error) {
	type alias Server
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "url")
	delete(x.Extensions, "description")
	delete(x.Extensions, "variables")
	*s = Server(x)
	return
}

func (s *Server) Validate() error {
	if s.URL == "" {
		return verifyError("url", fmt.Errorf("must be a non empty string"))
	}

	for k, v := range s.Variables {
		if err := v.Validate(); err != nil {
			return verifyError(fmt.Sprintf("variables[%v]", k), err)
		}
	}

	if s.Extensions != nil {
		if err := validatorExtensions(s.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options are from a limited set.
	// The array MUST NOT be empty.
	Enum []string `json:"enum"`

	// REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value
	// is not supplied. Note this behavior is different than the Schema Object's treatment of default
	// values, because in those cases parameter values are optional. If the enum is defined, the value
	// MUST exist in the enum's values.
	Default string `json:"default"`

	// An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (s *ServerVariable) marshalField() []marshalField {
	return []marshalField{
		{"enum", s.Enum, s.Enum == nil},
		{"default", s.Default, s.Default == ""},
		{"description", s.Description, s.Description == ""},
	}
}

func (s *ServerVariable) MarshalJSON() ([]byte, error) {
	return marshalJson(s.marshalField(), s.Extensions)
}

func (s *ServerVariable) UnmarshalJSON(buf []byte) (err error) {
	type alias ServerVariable
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "enum")
	delete(x.Extensions, "default")
	delete(x.Extensions, "description")
	*s = ServerVariable(x)
	return
}

func (s *ServerVariable) Validate() error {
	if s.Default == "" {
		return verifyError("default", fmt.Errorf("must be a non empty string"))
	}

	if s.Extensions != nil {
		if err := validatorExtensions(s.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Components struct {
	// An object to hold reusable Schema Objects.
	Schemas map[string]*Schema `json:"schemas"`

	// An object to hold reusable Response Objects.
	Responses map[string]*Response `json:"responses"`

	// An object to hold reusable Parameter Objects.
	Parameters map[string]*Parameter `json:"parameters"`

	// An object to hold reusable Example Objects.
	Examples map[string]*Example `json:"examples"`

	// An object to hold reusable Request Body Objects.
	RequestBodies map[string]*RequestBody `json:"requestBodies"`

	// An object to hold reusable Header Objects.
	Headers map[string]*Header `json:"headers"`

	// An object to hold reusable Security Scheme Objects.
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes"`

	// An object to hold reusable Link Objects.
	Links map[string]*Link `json:"links"`

	// An object to hold reusable Callback Objects.
	Callbacks map[string]*Callback `json:"callbacks"`

	// An object to hold reusable Path Item Object.
	PathItems map[string]*PathItem `json:"pathItems"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (c *Components) marshalField() []marshalField {
	return []marshalField{
		{"schemas", c.Schemas, c.Schemas == nil},
		{"responses", c.Responses, c.Responses == nil},
		{"parameters", c.Parameters, c.Parameters == nil},
		{"examples", c.Examples, c.Examples == nil},
		{"requestBodies", c.RequestBodies, c.RequestBodies == nil},
		{"headers", c.Headers, c.Headers == nil},
		{"securitySchemes", c.SecuritySchemes, c.SecuritySchemes == nil},
		{"links", c.Links, c.Links == nil},
		{"callbacks", c.Callbacks, c.Callbacks == nil},
		{"pathItems", c.PathItems, c.PathItems == nil},
	}
}

func (c *Components) MarshalJSON() ([]byte, error) {
	return marshalJson(c.marshalField(), c.Extensions)
}

func (c *Components) UnmarshalJSON(buf []byte) (err error) {
	type alias Components
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "schemas")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "requestBodies")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "securitySchemes")
	delete(x.Extensions, "links")
	delete(x.Extensions, "callbacks")
	delete(x.Extensions, "pathItems")
	*c = Components(x)
	return
}

func (c *Components) Validate(openapi *OpenAPI) error {
	for k, v := range c.Schemas {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("schemas[%v]", k), err)
		}
	}

	for k, v := range c.Responses {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("responses[%v]", k), err)
		}
	}

	for k, v := range c.Parameters {
		if err := v.Validate(openapi, ""); err != nil {
			return verifyError(fmt.Sprintf("parameters[%v]", k), err)
		}
	}

	for k, v := range c.Examples {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("examples[%v]", k), err)
		}
	}

	for k, v := range c.RequestBodies {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("requestBodies[%v]", k), err)
		}
	}

	for k, v := range c.Headers {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("headers[%v]", k), err)
		}
	}

	for k, v := range c.SecuritySchemes {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("securitySchemes[%v]", k), err)
		}
	}

	for k, v := range c.Links {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("links[%v]", k), err)
		}
	}

	for k, v := range c.Callbacks {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("callbacks[%v]", k), err, true)
		}
	}

	for k, v := range c.PathItems {
		if err := v.Validate(openapi, k); err != nil {
			return verifyError(fmt.Sprintf("pathItems[%v]", k), err)
		}
	}

	if c.Extensions != nil {
		if err := validatorExtensions(c.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Paths struct {
	m map[string]*PathItem

	Extensions map[string]any
}

func (p *Paths) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	for k, v := range p.Extensions {
		m[k] = v
	}
	for k, v := range p.m {
		m[k] = v
	}
	return packageJsonByMap(m)
}

func (p *Paths) UnmarshalJSON(buf []byte) (err error) {
	type alias Paths
	var m map[string]any
	if err = json.Unmarshal(buf, &m); err != nil {
		return
	}
	x := alias{
		m:          map[string]*PathItem{},
		Extensions: map[string]any{},
	}
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}
		var b []byte
		if b, err = json.Marshal(v); err != nil {
			return
		}
		var pathItem PathItem
		if err = json.Unmarshal(b, &pathItem); err != nil {
			return
		}
		x.m[k] = &pathItem
	}
	*p = Paths(x)
	return
}

func (p *Paths) Validate(openapi *OpenAPI) error {
	handlePaths := map[string]bool{}
	for k, v := range p.m {
		if k[0] != '/' {
			return verifyError(k, fmt.Errorf("key must start with \"/\""))
		}
		hPath, err := handlePath(k)
		if err != nil {
			return verifyError(k, err)
		}
		if handlePaths[hPath] {
			return verifyError(k, fmt.Errorf("path duplication"))
		}
		handlePaths[hPath] = true
		if err = v.Validate(openapi, k); err != nil {
			return verifyError(k, err)
		}
	}

	if p.Extensions != nil {
		if err := validatorExtensions(p.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

// Set A relative path to an individual endpoint. The field name MUST begin with a forward slash ().
// The path is appended (no relative URL resolution) to the expanded URL from the Server Object's
// field in order to construct the full URL. Path templating is allowed. When matching URLs,
// concrete (non-templated) paths would be matched before their templated counterparts. Templated
// paths with the same hierarchy but different templated names MUST NOT exist as they are identical.
// In case of ambiguous matching, it's up to the tooling to decide which one to use./url
func (p *Paths) Set(path string, item *PathItem) {
	if p.m == nil {
		p.m = map[string]*PathItem{}
	}
	p.m[path] = item
}

func (p *Paths) Value(path string) *PathItem {
	return p.m[path]
}

type PathItem struct {
	Ref string `json:"$ref"`

	// An optional, string summary, intended to apply to all operations in this path.
	Summary string `json:"summary"`

	// An optional, string description, intended to apply to all operations in this path.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// A definition of a GET operation on this path.
	Get *Operation `json:"get"`

	// A definition of a PUT operation on this path.
	Put *Operation `json:"put"`

	// A definition of a POST operation on this path.
	Post *Operation `json:"post"`

	// A definition of a DELETE operation on this path.
	Delete *Operation `json:"delete"`

	// A definition of a OPTIONS operation on this path.
	Options *Operation `json:"options"`

	// A definition of a HEAD operation on this path.
	Head *Operation `json:"head"`

	// A definition of a PATCH operation on this path.
	Patch *Operation `json:"patch"`

	// A definition of a TRACE operation on this path.
	Trace *Operation `json:"trace"`

	// An alternative array to service all operations in this path.server
	Servers []*Server `json:"servers"`

	// A list of parameters that are applicable for all the operations described under this path.
	// These parameters can be overridden at the operation level, but cannot be removed there. The
	// list MUST NOT include duplicated parameters. A unique parameter is defined by a combination
	// of a name and location. The list can use the Reference Object to link to parameters that are
	// defined at the OpenAPI Object's components/parameters.
	Parameters []*Parameter `json:"parameters"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (p *PathItem) marshalField() []marshalField {
	if p.Ref != "" {
		return []marshalField{
			{"$ref", p.Ref, false},
			{"summary", p.Summary, p.Summary == ""},
			{"description", p.Description, p.Description == ""},
		}
	}
	return []marshalField{
		{"summary", p.Summary, p.Summary == ""},
		{"description", p.Description, p.Description == ""},
		{"get", p.Get, p.Get == nil},
		{"put", p.Put, p.Put == nil},
		{"post", p.Post, p.Post == nil},
		{"delete", p.Delete, p.Delete == nil},
		{"options", p.Options, p.Options == nil},
		{"head", p.Head, p.Head == nil},
		{"patch", p.Patch, p.Patch == nil},
		{"trace", p.Trace, p.Trace == nil},
		{"servers", p.Servers, p.Servers == nil},
		{"parameters", p.Parameters, p.Parameters == nil},
	}
}

func (p *PathItem) MarshalJSON() ([]byte, error) {
	return marshalJson(p.marshalField(), p.Extensions)
}

func (p *PathItem) UnmarshalJSON(buf []byte) (err error) {
	type alias PathItem
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "get")
	delete(x.Extensions, "put")
	delete(x.Extensions, "post")
	delete(x.Extensions, "delete")
	delete(x.Extensions, "options")
	delete(x.Extensions, "head")
	delete(x.Extensions, "patch")
	delete(x.Extensions, "trace")
	delete(x.Extensions, "servers")
	delete(x.Extensions, "parameters")
	*p = PathItem(x)
	return
}

func (p *PathItem) Validate(openapi *OpenAPI, path string) error {
	if p.Ref != "" {
		if err := validatorRef(p.Ref, "pathItem", openapi); err != nil {
			return err
		}
		return nil
	}

	if p.Get != nil {
		if err := p.Get.Validate(openapi, path); err != nil {
			return verifyError("get", err)
		}
	}

	if p.Put != nil {
		if err := p.Put.Validate(openapi, path); err != nil {
			return verifyError("put", err)
		}
	}

	if p.Post != nil {
		if err := p.Post.Validate(openapi, path); err != nil {
			return verifyError("post", err)
		}
	}

	if p.Delete != nil {
		if err := p.Delete.Validate(openapi, path); err != nil {
			return verifyError("delete", err)
		}
	}

	if p.Options != nil {
		if err := p.Options.Validate(openapi, path); err != nil {
			return verifyError("options", err)
		}
	}

	if p.Head != nil {
		if err := p.Head.Validate(openapi, path); err != nil {
			return verifyError("head", err)
		}
	}

	if p.Patch != nil {
		if err := p.Patch.Validate(openapi, path); err != nil {
			return verifyError("patch", err)
		}
	}

	if p.Trace != nil {
		if err := p.Trace.Validate(openapi, path); err != nil {
			return verifyError("trace", err)
		}
	}

	for k, v := range p.Servers {
		if err := v.Validate(); err != nil {
			return verifyError(fmt.Sprintf("servers[%v]", k), err)
		}
	}

	for k, v := range p.Parameters {
		if err := v.Validate(openapi, path); err != nil {
			return verifyError(fmt.Sprintf("parameters[%v]", k), err)
		}
	}

	if p.Extensions != nil {
		if err := validatorExtensions(p.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Operation struct {
	// A list of tags for API documentation control. Tags can be used for logical grouping of
	// operations by resources or any other qualifier.
	Tags []string `json:"tags"`

	// A short summary of what the operation does.
	Summary string `json:"summary"`

	// A verbose explanation of the operation behavior. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// Additional external documentation for this operation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs"`

	// Unique string used to identify the operation. The id MUST be unique among all operations described in
	// the API. The operationId value is case-sensitive. Tools and libraries MAY use the operationId to uniquely
	// identify an operation, therefore, it is RECOMMENDED to follow common programming naming conventions.
	OperationId string `json:"operationId"`

	// A list of parameters that are applicable for this operation. If a parameter is already defined at the Path
	// Item, the new definition will override it but can never remove it. The list MUST NOT include duplicated
	// parameters. A unique parameter is defined by a combination of a name and location. The list can use
	// the Reference Object to link to parameters that are defined at the OpenAPI Object's components/parameters.
	Parameters []*Parameter `json:"parameters"`

	// The request body applicable for this operation. The is fully supported in HTTP methods where
	// the HTTP 1.1 specification RFC7231 has explicitly defined semantics for request bodies. In other
	// cases where the HTTP spec is vague (such as GET, HEAD and DELETE), is permitted but does not have
	// well-defined semantics and SHOULD be avoided if possible.requestBodyrequestBody
	RequestBody *RequestBody `json:"requestBody"`

	// The list of possible responses as they are returned from executing this operation.
	Responses *Responses `json:"responses"`

	// A map of possible out-of band callbacks related to the parent operation. The key is a unique
	// identifier for the Callback Object. Each value in the map is a Callback Object that describes a
	// request that may be initiated by the API provider and the expected responses.
	Callbacks map[string]*Callback `json:"callbacks"`

	// Declares this operation to be deprecated. Consumers SHOULD refrain from usage of the
	// declared operation. Default value is .false
	Deprecated bool `json:"deprecated"`

	// A declaration of which security mechanisms can be used for this operation. The list of values
	// includes alternative security requirement objects that can be used. Only one of the security
	// requirement objects need to be satisfied to authorize a request. To make security optional, an
	// empty security requirement () can be included in the array. This definition overrides any declared
	// top-level security. To remove a top-level security declaration, an empty array can be used.{}
	Security []*SecurityRequirement `json:"security"`

	// An alternative array to service this operation. If an alternative object is specified at the
	// Path Item Object or Root level, it will be overridden by this value.serverserver
	Servers []*Server `json:"servers"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (o *Operation) marshalField() []marshalField {
	return []marshalField{
		{"tags", o.Tags, o.Tags == nil},
		{"summary", o.Summary, o.Summary == ""},
		{"description", o.Description, o.Description == ""},
		{"externalDocs", o.ExternalDocs, o.ExternalDocs == nil},
		{"operationId", o.OperationId, o.OperationId == ""},
		{"parameters", o.Parameters, o.Parameters == nil},
		{"requestBody", o.RequestBody, o.RequestBody == nil},
		{"responses", o.Responses, o.Responses == nil},
		{"callbacks", o.Callbacks, o.Callbacks == nil},
		{"deprecated", o.Deprecated, o.Deprecated == false},
		{"security", o.Security, o.Security == nil},
		{"servers", o.Servers, o.Servers == nil},
	}
}

func (o *Operation) MarshalJSON() ([]byte, error) {
	return marshalJson(o.marshalField(), o.Extensions)
}

func (o *Operation) UnmarshalJSON(buf []byte) (err error) {
	type alias Operation
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "tags")
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "operationId")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "requestBody")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "callbacks")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "security")
	delete(x.Extensions, "servers")
	*o = Operation(x)
	return
}

func (o *Operation) Validate(openapi *OpenAPI, path string) error {
	if o.ExternalDocs != nil {
		if err := o.ExternalDocs.Validate(); err != nil {
			return verifyError("externalDocs", err)
		}
	}

	idMap := map[string]int{}
	for _, v := range openapi.Paths.m {
		if v.Get != nil {
			idMap[v.Get.OperationId]++
		}
		if v.Put != nil {
			idMap[v.Put.OperationId]++
		}
		if v.Post != nil {
			idMap[v.Post.OperationId]++
		}
		if v.Delete != nil {
			idMap[v.Delete.OperationId]++
		}
		if v.Options != nil {
			idMap[v.Options.OperationId]++
		}
		if v.Head != nil {
			idMap[v.Head.OperationId]++
		}
		if v.Patch != nil {
			idMap[v.Patch.OperationId]++
		}
		if v.Trace != nil {
			idMap[v.Trace.OperationId]++
		}
	}
	for _, c := range idMap {
		if c > 1 {
			return verifyError("operationId", fmt.Errorf("value duplication"))
		}
	}

	pathTotal := strings.Count(path, "{")
	pathCount := 0
	for k, v := range o.Parameters {
		if v.In == "path" {
			pathCount++
		}
		if err := v.Validate(openapi, path); err != nil {
			return verifyError(fmt.Sprintf("parameters[%v]", k), err)
		}
	}
	if pathCount != pathTotal {
		return verifyError("parameters", fmt.Errorf("must be in %q when in is \"path\"", path))
	}

	if o.RequestBody != nil {
		if err := o.RequestBody.Validate(openapi); err != nil {
			return verifyError("requestBody", err)
		}
	}

	if o.Responses != nil {
		if err := o.Responses.Validate(openapi); err != nil {
			return verifyError("responses", err, true)
		}
	}

	for k, v := range o.Callbacks {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("callbacks[%v]", k), err, true)
		}
	}

	for k, v := range o.Servers {
		if err := v.Validate(); err != nil {
			return verifyError(fmt.Sprintf("servers[%v]", k), err)
		}
	}

	if o.Extensions != nil {
		if err := validatorExtensions(o.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type ExternalDocumentation struct {
	// A description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// REQUIRED. The URL for the target documentation. This MUST be in the form of a URL.
	URL string `json:"url"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (e *ExternalDocumentation) marshalField() []marshalField {
	return []marshalField{
		{"description", e.Description, e.Description == ""},
		{"url", e.URL, e.URL == ""},
	}
}

func (e *ExternalDocumentation) MarshalJSON() ([]byte, error) {
	return marshalJson(e.marshalField(), e.Extensions)
}

func (e *ExternalDocumentation) UnmarshalJSON(buf []byte) (err error) {
	type alias ExternalDocumentation
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "description")
	delete(x.Extensions, "url")
	*e = ExternalDocumentation(x)
	return
}

func (e *ExternalDocumentation) Validate() error {
	if e.URL == "" {
		return verifyError("url", fmt.Errorf("must be a non empty string"))
	}

	if e.Extensions != nil {
		if err := validatorExtensions(e.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Parameter struct {
	Ref string `json:"$ref"`

	// REQUIRED. The name of the parameter. Parameter names are case sensitive.
	//   If in is "path", the name field MUST correspond to a template expression occurring within
	//     the path field in the Paths Object. See Path Templating for further information.
	//   If in is "header" and the name field is "Accept", "Content-Type" or "Authorization",
	//     the parameter definition SHALL be ignored.
	//   For all other cases, the name corresponds to the parameter name used by the in property.
	Name string `json:"name"`

	// REQUIRED. The location of the parameter. Possible values are , , or ."query""header""path""cookie"
	In string `json:"in"`

	// A brief description of the parameter. This could contain examples of use. CommonMark syntax
	// MAY be used for rich text representation.
	Description string `json:"description"`

	// Determines whether this parameter is mandatory. If the parameter location is "path", this
	// property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included
	// and its default value is false.
	Required bool `json:"required"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is .false
	Deprecated bool `json:"deprecated"`

	// Sets the ability to pass empty-valued parameters. This is valid only for parameters and allows
	// sending a parameter with an empty value. Default value is . If style is used, and if behavior is
	// (cannot be serialized), the value of SHALL be ignored. Use of this property is NOT RECOMMENDED,
	// as it is likely to be removed in a later revision.queryfalsen/aallowEmptyValue
	AllowEmptyValue bool `json:"allowEmptyValue"`

	// Describes how the parameter value will be serialized depending on the type of the parameter value. Default
	// values (based on value of in): for query - form; for path - simple; for header - simple; for cookie - form.
	Style string `json:"style"`

	// When this is true, parameter values of type or generate separate parameters for each value of
	// the array or key-value pair of the map. For other types of parameters this property has no
	// effect. When style is , the default value is . For all other styles, the default value is
	// .array object form true false
	Explode bool `json:"explode"`

	// Determines whether the parameter value SHOULD allow reserved characters, as defined by RFC3986
	// to be included without percent-encoding. This property only applies to parameters with an
	// value of . The default value is .:/?#[]@!$&'()*+,;=in query false
	AllowReserved bool `json:"allowReserved"`

	// The schema defining the type used for the parameter.
	Schema *Schema `json:"schema"`

	// Example of the parameter's potential value. The example SHOULD match the specified schema and
	// encoding properties if present. The field is mutually exclusive of the field. Furthermore, if
	// referencing a that contains an example, the value SHALL override the example provided by the
	// schema. To represent examples of media types that cannot naturally be represented in JSON or YAML,
	// a string value can contain the example with escaping where necessary.example examples schema example
	Example any `json:"example"`
	// Examples of the parameter's potential value. Each example SHOULD contain a value in the correct
	// format as specified in the parameter encoding. The field is mutually exclusive of the field.
	// Furthermore, if referencing a that contains an example, the value SHALL override the example
	// provided by the schema.examples example schema examples
	Examples map[string]*Example `json:"examples"`

	// A map containing the representations for the parameter. The key is the media type and the
	// value describes it. The map MUST only contain one entry.
	Content map[string]*MediaType `json:"content"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (p *Parameter) marshalField() []marshalField {
	if p.Ref != "" {
		return []marshalField{
			{"$ref", p.Ref, false},
			{"description", p.Description, p.Description == ""},
		}
	}
	if p.In == "header" && (p.Name == "Accept" || p.Name == "Content-Type" || p.Name == "Authorization") {
		p.Name = ""
	}
	return []marshalField{
		{"name", p.Name, p.Name == ""},
		{"in", p.In, p.In == ""},
		{"description", p.Description, p.Description == ""},
		{"required", p.Required, p.Required == false},
		{"deprecated", p.Deprecated, p.Deprecated == false},
		{"allowEmptyValue", p.AllowEmptyValue, p.AllowEmptyValue == false},
		{"style", p.Style, p.Style == ""},
		{"explode", p.Explode, p.Explode == false},
		{"allowReserved", p.AllowReserved, p.AllowReserved == false},
		{"schema", p.Schema, p.Schema == nil},
		{"example", p.Example, p.Example == nil},
		{"examples", p.Examples, p.Examples == nil},
		{"content", p.Content, p.Content == nil},
	}
}

func (p *Parameter) MarshalJSON() ([]byte, error) {
	return marshalJson(p.marshalField(), p.Extensions)
}

func (p *Parameter) UnmarshalJSON(buf []byte) (err error) {
	type alias Parameter
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "name")
	delete(x.Extensions, "in")
	delete(x.Extensions, "description")
	delete(x.Extensions, "required")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "style")
	delete(x.Extensions, "explode")
	delete(x.Extensions, "allowReserved")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "example")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "content")
	*p = Parameter(x)
	return
}

func (p *Parameter) Validate(openapi *OpenAPI, path string) error {
	if p.Ref != "" {
		if err := validatorRef(p.Ref, "parameter", openapi); err != nil {
			return err
		}
		return nil
	}

	if p.Name == "" {
		return verifyError("name", fmt.Errorf("must be a non empty string"))
	}

	if p.In == "path" && path != "" && !strings.Contains(path, p.Name) {
		return verifyError("name", fmt.Errorf("must be in %q when in is \"path\"", path))
	}

	if p.In == "" {
		return verifyError("in", fmt.Errorf("must be a non empty string"))
	}

	if p.In != "query" && p.In != "header" && p.In != "path" && p.In != "cookie" {
		return verifyError("in", fmt.Errorf("must be within \"query\", \"header\", \"path\", \"cookie\""))
	}

	if p.In == "path" && !p.Required {
		return verifyError("required", fmt.Errorf("must be \"true\" when in is \"path\""))
	}

	if p.Schema != nil {
		if err := p.Schema.Validate(openapi); err != nil {
			return verifyError("schema", err)
		}
	}

	if p.Example != nil && p.Examples != nil {
		return fmt.Errorf("fields example and examples are mutually exclusive")
	}

	for k, v := range p.Examples {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("examples[%v]", k), err)
		}
	}

	for k, v := range p.Content {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("content[%v]", k), err)
		}
	}

	if p.Extensions != nil {
		if err := validatorExtensions(p.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type RequestBody struct {
	Ref string `json:"$ref"`

	// A brief description of the request body. This could contain examples of use. CommonMark
	// syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// REQUIRED. The content of the request body. The key is a media type or media type range and the
	// value describes it. For requests that match multiple keys, only the most specific key is
	// applicable. e.g. text/plain overrides text/*
	Content map[string]*MediaType `json:"content"`

	// Determines if the request body is required in the request. Defaults to .false
	Required bool `json:"required"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (r *RequestBody) marshalField() []marshalField {
	if r.Ref != "" {
		return []marshalField{
			{"$ref", r.Ref, false},
			{"description", r.Description, r.Description == ""},
		}
	}
	return []marshalField{
		{"description", r.Description, r.Description == ""},
		{"content", r.Content, r.Content == nil},
		{"required", r.Required, r.Required == false},
	}
}

func (r *RequestBody) MarshalJSON() ([]byte, error) {
	return marshalJson(r.marshalField(), r.Extensions)
}

func (r *RequestBody) UnmarshalJSON(buf []byte) (err error) {
	type alias RequestBody
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "content")
	delete(x.Extensions, "required")
	*r = RequestBody(x)
	return
}

func (r *RequestBody) Validate(openapi *OpenAPI) error {
	if r.Ref != "" {
		if err := validatorRef(r.Ref, "requestBody", openapi); err != nil {
			return err
		}
		return nil
	}

	if r.Content != nil {
		for k, v := range r.Content {
			if err := v.Validate(openapi); err != nil {
				return verifyError(fmt.Sprintf("content[%v]", k), err)
			}
		}
	} else {
		return verifyError("content", fmt.Errorf("must be a non empty object"))
	}

	if r.Extensions != nil {
		if err := validatorExtensions(r.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type MediaType struct {
	// The schema defining the content of the request, response, or parameter.
	Schema *Schema `json:"schema"`

	// Example of the media type. The example object SHOULD be in the correct format as specified by the media
	// type. The field is mutually exclusive of the field. Furthermore, if referencing a which contains an
	// example, the value SHALL override the example provided by the schema.example examples schema example
	Example any `json:"example"`

	// Examples of the media type. Each example object SHOULD match the media type and specified schema if
	// present. The field is mutually exclusive of the field. Furthermore, if referencing a which contains
	// an example, the value SHALL override the example provided by the schema.examples example schema examples
	Examples map[string]*Example `json:"examples"`

	// A map between a property name and its encoding information. The key, being the property name,
	// MUST exist in the schema as a property. The encoding object SHALL only apply to objects when the
	// media type is or .requestBody multipart application/x-www-form-urlencoded
	Encoding map[string]*Encoding `json:"encoding"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (m *MediaType) marshalField() []marshalField {
	return []marshalField{
		{"schema", m.Schema, m.Schema == nil},
		{"example", m.Example, m.Example == nil},
		{"examples", m.Examples, m.Examples == nil},
		{"encoding", m.Encoding, m.Encoding == nil},
	}
}

func (m *MediaType) MarshalJSON() ([]byte, error) {
	return marshalJson(m.marshalField(), m.Extensions)
}

func (m *MediaType) UnmarshalJSON(buf []byte) (err error) {
	type alias MediaType
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "schema")
	delete(x.Extensions, "example")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "encoding")
	*m = MediaType(x)
	return
}

func (m *MediaType) Validate(openapi *OpenAPI) error {
	if m.Schema != nil {
		if err := m.Schema.Validate(openapi); err != nil {
			return verifyError("schema", err)
		}
	}

	if m.Example != nil && m.Examples != nil {
		return fmt.Errorf("fields example and examples are mutually exclusive")
	}

	for k, v := range m.Examples {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("examples[%v]", k), err)
		}
	}

	for k, v := range m.Encoding {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("encoding[%v]", k), err)
		}
	}

	if m.Extensions != nil {
		if err := validatorExtensions(m.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Encoding struct {
	// The Content-Type for encoding a specific property. Default value depends on the property type: for - ;
	// for – the default is defined based on the inner type; for all other cases the default  is . The value
	// can be a specific media type (e.g. ), a wildcard media type (e.g. ), or a comma-separated list of
	// the two types.object application/json array application/octet-stream application/json image/*
	ContentType string `json:"contentType"`

	// A map allowing additional information to be provided as headers, for example . is described separately
	// and SHALL be ignored in this section. This property SHALL be ignored if the request body media
	// type is not a .Content-Disposition Content-Type multipart
	Headers map[string]*Header `json:"headers"`

	// Describes how a specific property value will be serialized depending on its type. See Parameter
	// Object for details on the style property. The behavior follows the same values as parameters,
	// including default values. This property SHALL be ignored if the request body media type is not or .
	// If a value is explicitly defined, then the value of contentType (implicit or explicit) SHALL be
	// ignored .query application/x-www-form-urlencoded multipart/form-data
	Style string `json:"style"`

	// When this is true, property values of type or generate separate parameters for each value of the
	// array, or key-value-pair of the map. For other types of properties this property has no effect.
	// When style is , the default value is . For all other styles, the default value is . This property
	// SHALL be ignored if the request body media type is not or . If a value is explicitly defined, then
	// the value of contentType (implicit or explicit) SHALL be ignored .array object form true false
	// application/x-www-form-urlencoded multipart/form-data
	Explode bool `json:"explode"`

	// Determines whether the parameter value SHOULD allow reserved characters, as defined by RFC3986
	// to be included without percent-encoding. The default value is . This property SHALL be ignored if
	// the request body media type is not or . If a value is explicitly defined, then the value of
	// contentType (implicit or explicit) SHALL be ignored.:/?#[]@!$&'()*+,;=false
	// application/x-www-form-urlencoded multipart/form-data
	AllowReserved bool `json:"allowReserved"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (e *Encoding) marshalField() []marshalField {
	return []marshalField{
		{"contentType", e.ContentType, e.ContentType == ""},
		{"headers", e.Headers, e.Headers == nil},
		{"style", e.Style, e.Style == ""},
		{"explode", e.Explode, e.Explode == false},
		{"allowReserved", e.AllowReserved, e.AllowReserved == false},
	}
}

func (e *Encoding) MarshalJSON() ([]byte, error) {
	return marshalJson(e.marshalField(), e.Extensions)
}

func (e *Encoding) UnmarshalJSON(buf []byte) (err error) {
	type alias Encoding
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "contentType")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "style")
	delete(x.Extensions, "explode")
	delete(x.Extensions, "allowReserved")
	*e = Encoding(x)
	return
}

func (e *Encoding) Validate(openapi *OpenAPI) error {
	for k, v := range e.Headers {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("headers[%v]", k), err)
		}
	}

	if e.Extensions != nil {
		if err := validatorExtensions(e.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Responses struct {
	m map[string]*Response

	// The documentation of responses other than the ones declared for specific HTTP response codes.
	// Use this field to cover undeclared responses.
	Default *Response `json:"default"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (r *Responses) marshalField() []marshalField {
	var list []marshalField
	for k, v := range r.m {
		list = append(list, marshalField{k, v, false})
	}
	list = append(list, marshalField{"default", r.Default, r.Default == nil})
	return list
}

func (r *Responses) MarshalJSON() ([]byte, error) {
	return marshalJson(r.marshalField(), r.Extensions)
}

func (r *Responses) UnmarshalJSON(buf []byte) (err error) {
	type alias Responses
	var m map[string]any
	if err = json.Unmarshal(buf, &m); err != nil {
		return
	}
	x := alias{
		m:          map[string]*Response{},
		Default:    nil,
		Extensions: map[string]any{},
	}
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}
		var b []byte
		if b, err = json.Marshal(v); err != nil {
			return
		}
		var response Response
		if err = json.Unmarshal(b, &response); err != nil {
			return
		}
		if k == "default" {
			x.Default = &response
			continue
		}
		x.m[k] = &response
	}
	*r = Responses(x)
	return
}

func (r *Responses) Validate(openapi *OpenAPI) error {
	if r.Default != nil {
		if err := r.Default.Validate(openapi); err != nil {
			return verifyError("default", err)
		}
	}

	for k, v := range r.m {
		if err := v.Validate(openapi); err != nil {
			return verifyError(k, err)
		}
	}

	if r.Extensions != nil {
		if err := validatorExtensions(r.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

// Set Any HTTP status code can be used as the property name, but only one property per code, to describe
// the expected response for that HTTP status code. This field MUST be enclosed in quotation marks (for
// example, "200") for compatibility between JSON and YAML. To define a range of response codes, this field
// MAY contain the uppercase wildcard character . For example, represents all response codes between . Only
// the following range definitions are allowed: , , , , and . If a response is defined using an explicit
// code, the explicit code definition takes precedence over the range definition for that
// code. X 2XX [200-299] 1XX 2XX 3XX 4XX 5XX
func (r *Responses) Set(status string, response *Response) {
	if r.m == nil {
		r.m = map[string]*Response{}
	}
	r.m[status] = response
}

func (r *Responses) Value(status string) *Response {
	return r.m[status]
}

func (r *Responses) Responses() map[string]*Response {
	if r.Default != nil {
		if r.m == nil {
			r.m = map[string]*Response{}
		}
		r.m["default"] = r.Default
	}
	return r.m
}

type Response struct {
	Ref string `json:"$ref"`

	// REQUIRED. A description of the response. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// Maps a header name to its definition. RFC7230 states header names are case insensitive.
	// If a response header is defined with the name , it SHALL be ignored."Content-Type"
	Headers map[string]*Header `json:"headers"`

	// A map containing descriptions of potential response payloads. The key is a media type or media type
	// range and the value describes it. For responses that match multiple keys, only the most specific
	// key is applicable. e.g. text/plain overrides text/*
	Content map[string]*MediaType `json:"content"`

	// A map of operations links that can be followed from the response. The key of the map is a short
	// name for the link, following the naming constraints of the names for Component Objects.
	Links map[string]*Link `json:"links"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (r *Response) marshalField() []marshalField {
	if r.Ref != "" {
		return []marshalField{
			{"$ref", r.Ref, false},
			{"description", r.Description, r.Description == ""},
		}
	}
	return []marshalField{
		{"description", r.Description, r.Description == ""},
		{"headers", r.Headers, r.Headers == nil},
		{"content", r.Content, r.Content == nil},
		{"links", r.Links, r.Links == nil},
	}
}

func (r *Response) MarshalJSON() ([]byte, error) {
	return marshalJson(r.marshalField(), r.Extensions)
}

func (r *Response) UnmarshalJSON(buf []byte) (err error) {
	type alias Response
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "content")
	delete(x.Extensions, "links")
	*r = Response(x)
	return
}

func (r *Response) Validate(openapi *OpenAPI) error {
	if r.Ref != "" {
		if err := validatorRef(r.Ref, "response", openapi); err != nil {
			return err
		}
		return nil
	}

	if r.Description == "" {
		return verifyError("description", fmt.Errorf("must be a non empty string"))
	}

	for k, v := range r.Headers {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("headers[%v]", k), err)
		}
	}

	for k, v := range r.Content {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("content[%v]", k), err)
		}
	}

	for k, v := range r.Links {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("links[%v]", k), err)
		}
	}

	if r.Extensions != nil {
		if err := validatorExtensions(r.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

// Callback The key that identifies the Path Item Object is a runtime expression that can be evaluated in the context of a runtime HTTP request/response to identify the URL to be used for the callback request. A simple example might be $request.body#/url. However, using a runtime expression the complete HTTP message can be accessed. This includes accessing any part of a body that a JSON Pointer RFC6901 can reference.
//
// For example, given the following HTTP request:
//
// POST /subscribe/myevent?queryUrl=https://clientdomain.com/stillrunning HTTP/1.1
// Host: example.org
// Content-Type: application/json
// Content-Length: 187
//
//	{
//	 "failedUrl" : "https://clientdomain.com/failed",
//	 "successUrls" : [
//	   "https://clientdomain.com/fast",
//	   "https://clientdomain.com/medium",
//	   "https://clientdomain.com/slow"
//	 ]
//	}
//
// 201 Created
// Location: https://example.org/subscription/1
// The following examples show how the various expressions evaluate, assuming the callback operation has a path parameter named eventType and a query parameter named queryUrl.
//
// Expression	Value
// $url	https://example.org/subscribe/myevent?queryUrl=https://clientdomain.com/stillrunning
// $method	POST
// $request.path.eventType	myevent
// $request.query.queryUrl	https://clientdomain.com/stillrunning
// $request.header.content-Type	application/json
// $request.body#/failedUrl	https://clientdomain.com/failed
// $request.body#/successUrls/2	https://clientdomain.com/medium
// $response.header.Location	https://example.org/subscription/1
// Callback Object Examples
// The following example uses the user provided queryUrl query string parameter to define the callback URL. This is an example of how to use a callback object to describe a WebHook callback that goes with the subscription operation to enable registering for the WebHook.
//
// myCallback:
//
//	'{$request.query.queryUrl}':
//	  post:
//	    requestBody:
//	      description: Callback payload
//	      content:
//	        'application/json':
//	          schema:
//	            $ref: '#/components/schemas/SomePayload'
//	    responses:
//	      '200':
//	        description: callback successfully processed
//
// The following example shows a callback where the server is hard-coded, but the query string parameters are populated from the id and email property in the request body.
//
// transactionCallback:
//
//	'http://notificationServer.com?transactionId={$request.body#/id}&email={$request.body#/email}':
//	  post:
//	    requestBody:
//	      description: Callback payload
//	      content:
//	        'application/json':
//	          schema:
//	            $ref: '#/components/schemas/SomePayload'
//	    responses:
//	      '200':
//	        description: callback successfully processed
type Callback struct {
	Ref string `json:"$ref"`
	m   map[string]*PathItem

	Extensions map[string]any
}

func (c *Callback) marshalField() []marshalField {
	if c.Ref != "" {
		return []marshalField{{"$ref", c.Ref, false}}
	}
	var list []marshalField
	for k, v := range c.m {
		list = append(list, marshalField{k, v, false})
	}
	return list
}

func (c *Callback) MarshalJSON() ([]byte, error) {
	return marshalJson(c.marshalField(), c.Extensions)
}

func (c *Callback) UnmarshalJSON(buf []byte) (err error) {
	type alias Callback
	var m map[string]any
	if err = json.Unmarshal(buf, &m); err != nil {
		return
	}
	ref, _ := m["$ref"].(string)
	x := alias{
		Ref:        ref,
		m:          map[string]*PathItem{},
		Extensions: map[string]any{},
	}
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}
		var b []byte
		if b, err = json.Marshal(v); err != nil {
			return
		}
		var pathItem PathItem
		if err = json.Unmarshal(b, &pathItem); err != nil {
			return
		}
		x.m[k] = &pathItem
	}
	*c = Callback(x)
	return
}

func (c *Callback) Validate(openapi *OpenAPI) error {
	if c.Ref != "" {
		if err := validatorRef(c.Ref, "callback", openapi); err != nil {
			return err
		}
		return nil
	}

	for k, v := range c.m {
		if err := v.Validate(openapi, k); err != nil {
			return verifyError(k, err)
		}
	}

	if c.Extensions != nil {
		if err := validatorExtensions(c.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

// Set A Path Item Object, or a reference to one, used to define a callback request and expected responses.
// A complete example is available.
func (c *Callback) Set(path string, item *PathItem) {
	if c.m == nil {
		c.m = map[string]*PathItem{}
	}
	c.m[path] = item
}

func (c *Callback) Value(path string) *PathItem {
	return c.m[path]
}

type Example struct {
	Ref string `json:"$ref"`

	// Short description for the example.
	Summary string `json:"summary"`

	// Long description for the example. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// Embedded literal example. The field and field are mutually exclusive. To represent examples of media
	// types that cannot naturally represented in JSON or YAML, use a string value to contain the example,
	// escaping where necessary. value externalValue
	Value any `json:"value"`

	// A URI that points to the literal example. This provides the capability to reference examples that
	// cannot easily be included in JSON or YAML documents. The field and field are mutually exclusive.
	// See the rules for resolving Relative References. value externalValue
	ExternalValue string `json:"externalValue"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (e *Example) marshalField() []marshalField {
	if e.Ref != "" {
		return []marshalField{
			{"$ref", e.Ref, false},
			{"summary", e.Summary, e.Summary == ""},
			{"description", e.Description, e.Description == ""},
		}
	}
	return []marshalField{
		{"summary", e.Summary, e.Summary == ""},
		{"description", e.Description, e.Description == ""},
		{"value", e.Value, e.Value == nil},
		{"externalValue", e.ExternalValue, e.ExternalValue == ""},
	}
}

func (e *Example) MarshalJSON() ([]byte, error) {
	return marshalJson(e.marshalField(), e.Extensions)
}

func (e *Example) UnmarshalJSON(buf []byte) (err error) {
	type alias Example
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "summary")
	delete(x.Extensions, "description")
	delete(x.Extensions, "value")
	delete(x.Extensions, "externalValue")
	*e = Example(x)
	return
}

func (e *Example) Validate(openapi *OpenAPI) error {
	if e.Ref != "" {
		if err := validatorRef(e.Ref, "example", openapi); err != nil {
			return err
		}
		return nil
	}

	if e.Extensions != nil {
		if err := validatorExtensions(e.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Link struct {
	Ref string `json:"$ref"`

	// A relative or absolute URI reference to an OAS operation. This field is mutually exclusive of
	// the field, and MUST point to an Operation Object. Relative values MAY be used to locate an
	// existing Operation Object in the OpenAPI definition. See the rules for resolving Relative
	// References. operationId operationRef
	OperationRef string `json:"operationRef"`

	// The name of an existing, resolvable OAS operation, as defined with a unique . This field is
	// mutually exclusive of the field. operationId operationRef
	OperationId string `json:"operationId"`
	// A map representing parameters to pass to an operation as specified with or identified via .
	// The key is the parameter name to be used, whereas the value can be a constant or an expression
	// to be evaluated and passed to the linked operation. The parameter name can be qualified using
	// the parameter location for operations that use the same parameter name in different locations
	// (e.g. path.id). operationId operationRef [{in}.]{name}
	Parameters map[string]any `json:"parameters"`

	// A literal value or {expression} to use as a request body when calling the target operation.
	RequestBody any `json:"requestBody"`

	// A description of the link. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// A server object to be used by the target operation.
	Server *Server `json:"server"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (l *Link) marshalField() []marshalField {
	if l.Ref != "" {
		return []marshalField{
			{"$ref", l.Ref, false},
			{"description", l.Description, l.Description == ""},
		}
	}
	return []marshalField{
		{"operationRef", l.OperationRef, l.OperationRef == ""},
		{"operationId", l.OperationId, l.OperationId == ""},
		{"parameters", l.Parameters, l.Parameters == nil},
		{"requestBody", l.RequestBody, l.RequestBody == nil},
		{"description", l.Description, l.Description == ""},
		{"server", l.Server, l.Server == nil},
	}
}

func (l *Link) MarshalJSON() ([]byte, error) {
	return marshalJson(l.marshalField(), l.Extensions)
}

func (l *Link) UnmarshalJSON(buf []byte) (err error) {
	type alias Link
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "operationRef")
	delete(x.Extensions, "operationId")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "requestBody")
	delete(x.Extensions, "description")
	delete(x.Extensions, "server")
	*l = Link(x)
	return
}

func (l *Link) Validate(openapi *OpenAPI) error {
	if l.Ref != "" {
		if err := validatorRef(l.Ref, "link", openapi); err != nil {
			return err
		}
		return nil
	}

	if l.OperationRef != "" && l.OperationId != "" {
		return fmt.Errorf("fields operationRef and operationId are mutually exclusive")
	}

	if l.OperationRef != "" {
		if err := validatorRef(l.Ref, "operation", openapi); err != nil {
			return err
		}
		return nil
	}

	if l.Server != nil {
		if err := l.Server.Validate(); err != nil {
			return verifyError("server", err)
		}
	}

	if l.Extensions != nil {
		if err := validatorExtensions(l.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

// The Header Object follows the structure of the Parameter Object with the following changes:
//
//  1. name MUST NOT be specified, it is given in the corresponding map.headers
//  2. in MUST NOT be specified, it is implicitly in .header
//  3. All traits that are affected by the location MUST be applicable to a location of (for example, style).header
type Header Parameter

func (h *Header) marshalField() []marshalField {
	if h.Ref != "" {
		return []marshalField{
			{"$ref", h.Ref, false},
			{"description", h.Description, h.Description == ""},
		}
	}
	return []marshalField{
		{"description", h.Description, h.Description == ""},
		{"required", h.Required, h.In == "path"},
		{"deprecated", h.Deprecated, h.Deprecated == false},
		{"allowEmptyValue", h.AllowEmptyValue, h.AllowEmptyValue == false},
		{"style", h.Style, h.Style == ""},
		{"explode", h.Explode, h.Explode == false},
		{"allowReserved", h.AllowReserved, h.AllowReserved == false},
		{"schema", h.Schema, h.Schema == nil},
		{"example", h.Example, h.Example == nil},
		{"examples", h.Examples, h.Examples == nil},
		{"content", h.Content, h.Content == nil},
	}
}

func (h *Header) MarshalJSON() ([]byte, error) {
	return marshalJson(h.marshalField(), h.Extensions)
}

func (h *Header) UnmarshalJSON(buf []byte) (err error) {
	type alias Header
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "description")
	delete(x.Extensions, "required")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "allowEmptyValue")
	delete(x.Extensions, "style")
	delete(x.Extensions, "explode")
	delete(x.Extensions, "allowReserved")
	delete(x.Extensions, "schema")
	delete(x.Extensions, "example")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "content")
	*h = Header(x)
	return
}

func (h *Header) Validate(openapi *OpenAPI) error {
	if h.Ref != "" {
		if err := validatorRef(h.Ref, "header", openapi); err != nil {
			return err
		}
		return nil
	}

	if h.Name != "" {
		return verifyError("name", fmt.Errorf("must not be specified"))
	}

	if h.In != "" {
		return verifyError("in", fmt.Errorf("must not be specified"))
	}

	if h.Schema != nil {
		if err := h.Schema.Validate(openapi); err != nil {
			return verifyError("schema", err)
		}
	}

	if h.Example != nil && h.Examples != nil {
		return fmt.Errorf("fields example and examples are mutually exclusive")
	}

	for k, v := range h.Examples {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("examples[%v]", k), err)
		}
	}

	for k, v := range h.Content {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("content[%v]", k), err)
		}
	}

	if h.Extensions != nil {
		if err := validatorExtensions(h.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Tag struct {
	// REQUIRED. The name of the tag.
	Name string `json:"name"`

	// A description for the tag. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// Additional external documentation for this tag.
	ExternalDocs *ExternalDocumentation `json:"externalDocs"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (t *Tag) marshalField() []marshalField {
	return []marshalField{
		{"name", t.Name, t.Name == ""},
		{"description", t.Description, t.Description == ""},
		{"externalDocs", t.ExternalDocs, t.ExternalDocs == nil},
	}
}

func (t *Tag) MarshalJSON() ([]byte, error) {
	return marshalJson(t.marshalField(), t.Extensions)
}

func (t *Tag) UnmarshalJSON(buf []byte) (err error) {
	type alias Tag
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "name")
	delete(x.Extensions, "description")
	delete(x.Extensions, "externalDocs")
	*t = Tag(x)
	return
}

func (t *Tag) Validate() error {
	if t.Name == "" {
		return verifyError("name", fmt.Errorf("must be a non empty string"))
	}

	if t.ExternalDocs != nil {
		if err := t.ExternalDocs.Validate(); err != nil {
			return verifyError("externalDocs", err)
		}
	}

	if t.Extensions != nil {
		if err := validatorExtensions(t.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Schema struct {
	Ref string `json:"$ref"`
	// json schema
	Type   string `json:"type"` // Value MUST be a string. Multiple types via an array are not supported.
	Format string `json:"format"`
	Enum   []any  `json:"enum"`
	Const  any    `json:"const"` // Use of this keyword is functionally equivalent to an "enum"
	// basic
	Title       string `json:"title"`
	Description string `json:"description"`
	Default     any    `json:"default"`
	Deprecated  bool   `json:"deprecated"`
	ReadOnly    bool   `json:"readOnly"`
	WriteOnly   bool   `json:"writeOnly"`
	Examples    []any  `json:"examples"`
	// number
	MultipleOf       *float64 `json:"multipleOf"`
	Maximum          *float64 `json:"maximum"`          // <=
	ExclusiveMaximum *float64 `json:"exclusiveMaximum"` // <
	Minimum          *float64 `json:"minimum"`          // >=
	ExclusiveMinimum *float64 `json:"exclusiveMinimum"` // >
	// string
	MaxLength        *uint64        `json:"maxLength"`
	MinLength        uint64         `json:"minLength"`
	Pattern          string         `json:"pattern"`
	ContentEncoding  string         `json:"contentEncoding"`
	ContentMediaType string         `json:"contentMediaType"`
	ContentSchema    map[string]any `json:"contentSchema"`
	// array
	Items       *Schema `json:"items"`
	MaxItems    *uint64 `json:"maxItems"`
	MinItems    uint64  `json:"minItems"`
	UniqueItems bool    `json:"uniqueItems"`
	MaxContains *uint64 `json:"maxContains"`
	MinContains *uint64 `json:"minContains"`
	// object
	Properties        map[string]*Schema  `json:"properties"`
	MaxProperties     *uint64             `json:"maxProperties"`
	MinProperties     uint64              `json:"minProperties"`
	Required          []string            `json:"required"`
	DependentRequired map[string][]string `json:"dependentRequired"`

	OneOf []*Schema `json:"oneOf"`
	AnyOf []*Schema `json:"anyOf"`
	AllOf []*Schema `json:"allOf"`
	Not   *Schema   `json:"not"`

	// Adds support for polymorphism. The discriminator is an object name that is used to differentiate between
	// other schemas which may satisfy the payload description. See Composition and Inheritance for more details.
	Discriminator *Discriminator `json:"discriminator"`

	// This MAY be used only on properties schemas. It has no effect on root schemas. Adds additional metadata
	// to describe the XML representation of this property.
	XML *XML `json:"xml"`

	// Additional external documentation for this schema.
	ExternalDocs *ExternalDocumentation `json:"externalDocs"`

	// A free-form property to include an example of an instance for this schema. To represent examples that
	// cannot be naturally represented in JSON or YAML, a string value can be used to contain the example with
	// escaping where necessary.
	// Deprecated: The property has been deprecated in favor of the JSON Schema keyword. Use of is discouraged,
	// and later versions of this specification may remove it. example examples example
	Example any `json:"example"`

	// This object MAY be extended with Specification Extensions, though as noted, additional properties
	// MAY omit the prefix within this object.x-
	Extensions map[string]any
}

func (s *Schema) marshalField() []marshalField {
	if s.Ref != "" {
		return []marshalField{
			{"$ref", s.Ref, false},
			{"description", s.Description, s.Description == ""},
			{"xml", s.XML, s.XML == nil},
		}
	}
	return []marshalField{
		{"type", s.Type, s.Type == ""},
		{"format", s.Format, s.Format == ""},
		{"enum", s.Enum, s.Enum == nil},
		{"const", s.Const, s.Const == nil},
		{"title", s.Title, s.Title == ""},
		{"description", s.Description, s.Description == ""},
		{"default", s.Default, s.Default == nil},
		{"deprecated", s.Deprecated, s.Deprecated == false},
		{"readOnly", s.ReadOnly, s.ReadOnly == false},
		{"writeOnly", s.WriteOnly, s.WriteOnly == false},
		{"examples", s.Examples, s.Examples == nil},
		{"multipleOf", s.MultipleOf, s.MultipleOf == nil},
		{"maximum", s.Maximum, s.Maximum == nil},
		{"exclusiveMaximum", s.ExclusiveMaximum, s.ExclusiveMaximum == nil},
		{"minimum", s.Minimum, s.Minimum == nil},
		{"exclusiveMinimum", s.ExclusiveMinimum, s.ExclusiveMinimum == nil},
		{"maxLength", s.MaxLength, s.MaxLength == nil},
		{"minLength", s.MinLength, s.MinLength == 0},
		{"pattern", s.Pattern, s.Pattern == ""},
		{"contentEncoding", s.ContentEncoding, s.ContentEncoding == ""},
		{"contentMediaType", s.ContentMediaType, s.ContentMediaType == ""},
		{"contentSchema", s.ContentSchema, s.ContentSchema == nil},
		{"items", s.Items, s.Items == nil},
		{"maxItems", s.MaxItems, s.MaxItems == nil},
		{"minItems", s.MinItems, s.MinItems == 0},
		{"uniqueItems", s.UniqueItems, s.UniqueItems == false},
		{"maxContains", s.MaxContains, s.MaxContains == nil},
		{"minContains", s.MinContains, s.MinContains == nil},
		{"properties", s.Properties, s.Properties == nil},
		{"maxProperties", s.MaxProperties, s.MaxProperties == nil},
		{"minProperties", s.MinProperties, s.MinProperties == 0},
		{"required", s.Required, s.Required == nil},
		{"dependentRequired", s.DependentRequired, s.DependentRequired == nil},
		{"oneOf", s.OneOf, s.OneOf == nil},
		{"anyOf", s.AnyOf, s.AnyOf == nil},
		{"allOf", s.AllOf, s.AllOf == nil},
		{"not", s.Not, s.Not == nil},
		{"discriminator", s.Discriminator, s.Discriminator == nil},
		{"xml", s.XML, s.XML == nil},
		{"externalDocs", s.ExternalDocs, s.ExternalDocs == nil},
		{"example", s.Example, s.Example == nil},
	}
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	return marshalJson(s.marshalField(), s.Extensions)
}

func (s *Schema) UnmarshalJSON(buf []byte) (err error) {
	type alias Schema
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "type")
	delete(x.Extensions, "format")
	delete(x.Extensions, "enum")
	delete(x.Extensions, "const")
	delete(x.Extensions, "title")
	delete(x.Extensions, "description")
	delete(x.Extensions, "default")
	delete(x.Extensions, "deprecated")
	delete(x.Extensions, "readOnly")
	delete(x.Extensions, "writeOnly")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "multipleOf")
	delete(x.Extensions, "maximum")
	delete(x.Extensions, "exclusiveMaximum")
	delete(x.Extensions, "minimum")
	delete(x.Extensions, "exclusiveMinimum")
	delete(x.Extensions, "maxLength")
	delete(x.Extensions, "minLength")
	delete(x.Extensions, "pattern")
	delete(x.Extensions, "contentEncoding")
	delete(x.Extensions, "contentMediaType")
	delete(x.Extensions, "contentSchema")
	delete(x.Extensions, "items")
	delete(x.Extensions, "maxItems")
	delete(x.Extensions, "minItems")
	delete(x.Extensions, "uniqueItems")
	delete(x.Extensions, "maxContains")
	delete(x.Extensions, "minContains")
	delete(x.Extensions, "properties")
	delete(x.Extensions, "maxProperties")
	delete(x.Extensions, "minProperties")
	delete(x.Extensions, "required")
	delete(x.Extensions, "dependentRequired")
	delete(x.Extensions, "oneOf")
	delete(x.Extensions, "anyOf")
	delete(x.Extensions, "allOf")
	delete(x.Extensions, "not")
	delete(x.Extensions, "discriminator")
	delete(x.Extensions, "xml")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "example")
	*s = Schema(x)
	return
}

func (s *Schema) Validate(openapi *OpenAPI) error {
	if s.Ref != "" {
		if err := validatorRef(s.Ref, "schema", openapi); err != nil {
			return err
		}
		return nil
	}
	if s.Type == "" {
		return verifyError("type", fmt.Errorf("type must be a non empty string"))
	}

	switch s.Type {
	case "integer", "number":
	case "string":
	case "boolean":
	case "array":
	case "object":
	default:
		return verifyError("type", fmt.Errorf("must be within "+
			"\"integer\", \"number\", \"string\", \"boolean\", \"array\", \"object\""))
	}

	if s.Items != nil {
		if err := s.Items.Validate(openapi); err != nil {
			return verifyError("items", err)
		}
	}

	for k, v := range s.Properties {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("properties[%v]", k), err)
		}
	}

	for k, v := range s.OneOf {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("oneOf[%v]", k), err)
		}
	}

	for k, v := range s.AnyOf {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("anyOf[%v]", k), err)
		}
	}

	for k, v := range s.AllOf {
		if err := v.Validate(openapi); err != nil {
			return verifyError(fmt.Sprintf("allOf[%v]", k), err)
		}
	}

	if s.Not != nil {
		if err := s.Not.Validate(openapi); err != nil {
			return verifyError("not", err)
		}
	}

	if s.Discriminator != nil {
		if err := s.Discriminator.Validate(); err != nil {
			return verifyError("discriminator", err)
		}
	}

	if s.XML != nil {
		if err := s.XML.Validate(); err != nil {
			return verifyError("xml", err)
		}
	}

	if s.ExternalDocs != nil {
		if err := s.ExternalDocs.Validate(); err != nil {
			return verifyError("externalDocs", err)
		}
	}

	if s.Extensions != nil {
		if err := validatorExtensions(s.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type Discriminator struct {
	// REQUIRED. The name of the property in the payload that will hold the discriminator value.
	PropertyName string `json:"propertyName"`

	// An object to hold mappings between payload values and schema names or references.
	Mapping map[string]string `json:"mapping"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (d *Discriminator) marshalField() []marshalField {
	return []marshalField{
		{"propertyName", d.PropertyName, d.PropertyName == ""},
		{"mapping", d.Mapping, d.Mapping == nil},
	}
}

func (d *Discriminator) MarshalJSON() ([]byte, error) {
	return marshalJson(d.marshalField(), d.Extensions)
}

func (d *Discriminator) UnmarshalJSON(buf []byte) (err error) {
	type alias Discriminator
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "propertyName")
	delete(x.Extensions, "mapping")
	*d = Discriminator(x)
	return
}

func (d *Discriminator) Validate() error {
	if d.PropertyName == "" {
		return verifyError("propertyName", fmt.Errorf("must be a non empty string"))
	}

	if d.Extensions != nil {
		if err := validatorExtensions(d.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type XML struct {
	// Replaces the name of the element/attribute used for the described schema property. When defined
	// within , it will affect the name of the individual XML elements within the list. When defined
	// alongside being (outside the ), it will affect the wrapping element and only if is . If is , it
	// will be ignored. items type array items wrapped true wrapped false
	Name string `json:"name"`

	// The URI of the namespace definition. This MUST be in the form of an absolute URI.
	Namespace string `json:"namespace"`

	// The prefix to be used for the name.
	Prefix string `json:"prefix"`

	// Declares whether the property definition translates to an attribute instead of an element.
	// Default value is. false
	Attribute bool `json:"attribute"`

	// MAY be used only for an array definition. Signifies whether the array is wrapped (for example, ) or
	// unwrapped (). Default value is . The definition takes effect only when defined alongside being
	// (outside the ). false type array items
	Wrapped bool `json:"wrapped"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (x *XML) marshalField() []marshalField {
	return []marshalField{
		{"name", x.Name, x.Name == ""},
		{"namespace", x.Namespace, x.Namespace == ""},
		{"prefix", x.Prefix, x.Prefix == ""},
		{"attribute", x.Attribute, x.Attribute == false},
		{"wrapped", x.Wrapped, x.Wrapped == false},
	}
}

func (x *XML) MarshalJSON() ([]byte, error) {
	return marshalJson(x.marshalField(), x.Extensions)
}

func (x *XML) UnmarshalJSON(buf []byte) (err error) {
	type alias XML
	var x1 alias
	if err = json.Unmarshal(buf, &x1); err != nil {
		return
	}
	delete(x1.Extensions, "name")
	delete(x1.Extensions, "namespace")
	delete(x1.Extensions, "prefix")
	delete(x1.Extensions, "attribute")
	delete(x1.Extensions, "wrapped")
	*x = XML(x1)
	return
}

func (x *XML) Validate() error {
	if x.Extensions != nil {
		if err := validatorExtensions(x.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type SecurityScheme struct {
	Ref string `json:"$ref"`

	// REQUIRED. The type of the security scheme. Valid values are "apiKey" "http" "mutualTLS" "oauth2" "openIdConnect"
	Type string `json:"type"`

	// A description for security scheme. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// REQUIRED. The name of the header, query or cookie parameter to be used.
	Name string `json:"name"`

	// REQUIRED. The location of the API key. Valid values are , or "query" "header" "cookie"
	In string `json:"in"`

	// REQUIRED. The name of the HTTP Authorization scheme to be used in the Authorization header as
	// defined in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.
	Scheme string `json:"scheme"`

	// A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually
	// generated by an authorization server, so this information is primarily for documentation purposes.
	BearerFormat string `json:"bearerFormat"`

	// REQUIRED. An object containing configuration information for the flow types supported.
	Flows *OAuthFlows `json:"flows"`

	// REQUIRED. OpenId Connect URL to discover OAuth2 configuration values. This MUST be in the
	// form of a URL. The OpenID Connect standard requires the use of TLS.
	OpenIdConnectUrl string `json:"openIdConnectUrl"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (s *SecurityScheme) marshalField() []marshalField {
	if s.Ref != "" {
		return []marshalField{
			{"$ref", s.Ref, false},
			{"description", s.Description, s.Description == ""},
		}
	}
	return []marshalField{
		{"type", s.Type, s.Type == ""},
		{"description", s.Description, s.Description == ""},
		{"name", s.Name, s.Name == ""},
		{"in", s.In, s.In == ""},
		{"scheme", s.Scheme, s.Scheme == ""},
		{"bearerFormat", s.BearerFormat, s.BearerFormat == ""},
		{"flows", s.Flows, s.Flows == nil},
		{"openIdConnectUrl", s.OpenIdConnectUrl, s.OpenIdConnectUrl == ""},
	}
}

func (s *SecurityScheme) MarshalJSON() ([]byte, error) {
	return marshalJson(s.marshalField(), s.Extensions)
}

func (s *SecurityScheme) UnmarshalJSON(buf []byte) (err error) {
	type alias SecurityScheme
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "$ref")
	delete(x.Extensions, "type")
	delete(x.Extensions, "description")
	delete(x.Extensions, "name")
	delete(x.Extensions, "in")
	delete(x.Extensions, "scheme")
	delete(x.Extensions, "bearerFormat")
	delete(x.Extensions, "flows")
	delete(x.Extensions, "openIdConnectUrl")
	*s = SecurityScheme(x)
	return
}

func (s *SecurityScheme) Validate(openapi *OpenAPI) error {
	if s.Ref != "" {
		if err := validatorRef(s.Ref, "securityScheme", openapi); err != nil {
			return err
		}
		return nil
	}

	if s.Type == "" {
		return verifyError("type", fmt.Errorf("must be a non empty string"))
	}

	switch s.Type {
	case "apiKey":
		if s.Name == "" {
			return verifyError("name", fmt.Errorf("must be a non empty string when type is %q", s.Type))
		}

		if s.In == "" {
			return verifyError("in", fmt.Errorf("must be a non empty string when type is %q", s.Type))
		}

		if s.In != "query" && s.In != "header" && s.In != "cookie" {
			return verifyError("in", fmt.Errorf("must be within \"query\", \"header\", \"cookie\""))
		}
	case "http":
		if s.Scheme == "" {
			return verifyError("scheme", fmt.Errorf("must be a non empty string when type is %q", s.Type))
		}
	case "mutualTLS":
	case "oauth2":
		if s.Flows == nil {
			return verifyError("flows", fmt.Errorf("must be a non empty object when type is %q", s.Type))
		}

		if err := s.Flows.Validate(); err != nil {
			return verifyError("flows", err)
		}
	case "openIdConnect":
		if s.OpenIdConnectUrl == "" {
			return verifyError("openIdConnectUrl", fmt.Errorf("must be a non empty string when type is %q", s.Type))
		}
	default:
		return verifyError("type", fmt.Errorf("must be within \"apiKey\", \"http\", \"mutualTLS\", "+
			"\"oauth2\", \"openIdConnect\""))
	}

	if s.Extensions != nil {
		if err := validatorExtensions(s.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type OAuthFlows struct {
	// Configuration for the OAuth Implicit flow
	Implicit *OAuthFlow `json:"implicit"`

	// Configuration for the OAuth Resource Owner Password flow
	Password *OAuthFlow `json:"password"`

	// Configuration for the OAuth Client Credentials flow. Previously called in OpenAPI 2.0. application
	ClientCredentials *OAuthFlow `json:"clientCredentials"`

	// Configuration for the OAuth Authorization Code flow. Previously called in OpenAPI 2.0. accessCode
	AuthorizationCode *OAuthFlow `json:"authorizationCode"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (o *OAuthFlows) marshalField() []marshalField {
	return []marshalField{
		{"implicit", o.Implicit, o.Implicit == nil},
		{"password", o.Password, o.Password == nil},
		{"clientCredentials", o.ClientCredentials, o.ClientCredentials == nil},
		{"authorizationCode", o.AuthorizationCode, o.AuthorizationCode == nil},
	}
}

func (o *OAuthFlows) MarshalJSON() ([]byte, error) {
	return marshalJson(o.marshalField(), o.Extensions)
}

func (o *OAuthFlows) UnmarshalJSON(buf []byte) (err error) {
	type alias OAuthFlows
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "implicit")
	delete(x.Extensions, "password")
	delete(x.Extensions, "clientCredentials")
	delete(x.Extensions, "authorizationCode")
	*o = OAuthFlows(x)
	return
}

func (o *OAuthFlows) Validate() error {
	if o.Implicit != nil {
		if err := o.Implicit.Validate("implicit"); err != nil {
			return verifyError("implicit", err)
		}
	}

	if o.Password != nil {
		if err := o.Password.Validate("password"); err != nil {
			return verifyError("password", err)
		}
	}

	if o.ClientCredentials != nil {
		if err := o.ClientCredentials.Validate("clientCredentials"); err != nil {
			return verifyError("clientCredentials", err)
		}
	}

	if o.AuthorizationCode != nil {
		if err := o.AuthorizationCode.Validate("authorizationCode"); err != nil {
			return verifyError("authorizationCode", err)
		}
	}

	if o.Extensions != nil {
		if err := validatorExtensions(o.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type OAuthFlow struct {
	// REQUIRED. The authorization URL to be used for this flow. This MUST be in the form of a URL.
	// The OAuth2 standard requires the use of TLS.
	AuthorizationUrl string `json:"authorizationUrl"`

	// REQUIRED. The token URL to be used for this flow. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	TokenUrl string `json:"tokenUrl"`

	// The URL to be used for obtaining refresh tokens. This MUST be in the form of a URL. The OAuth2
	// standard requires the use of TLS.
	RefreshUrl string `json:"refreshUrl"`

	// REQUIRED. The available scopes for the OAuth2 security scheme. A map between the scope name and a
	// short description for it. The map MAY be empty.
	Scopes map[string]string `json:"scopes"`

	// This object MAY be extended with Specification Extensions.
	Extensions map[string]any
}

func (o *OAuthFlow) marshalField() []marshalField {
	return []marshalField{
		{"authorizationUrl", o.AuthorizationUrl, o.AuthorizationUrl == ""},
		{"tokenUrl", o.TokenUrl, o.TokenUrl == ""},
		{"refreshUrl", o.RefreshUrl, o.RefreshUrl == ""},
		{"scopes", o.Scopes, o.Scopes == nil},
	}
}

func (o *OAuthFlow) MarshalJSON() ([]byte, error) {
	return marshalJson(o.marshalField(), o.Extensions)
}

func (o *OAuthFlow) UnmarshalJSON(buf []byte) (err error) {
	type alias OAuthFlow
	var x alias
	if err = json.Unmarshal(buf, &x); err != nil {
		return
	}
	delete(x.Extensions, "authorizationUrl")
	delete(x.Extensions, "tokenUrl")
	delete(x.Extensions, "refreshUrl")
	delete(x.Extensions, "scopes")
	*o = OAuthFlow(x)
	return
}

func (o *OAuthFlow) Validate(applyTo string) error {
	switch applyTo {
	case "implicit":
		if o.AuthorizationUrl == "" {
			return verifyError("authorizationUrl", fmt.Errorf("must be a non empty string "+
				"when the object is %v", applyTo))
		}
	case "password":
		if o.TokenUrl == "" {
			return verifyError("tokenUrl", fmt.Errorf("must be a non empty string "+
				"when the object is %v", applyTo))
		}
	case "clientCredentials":
		if o.TokenUrl == "" {
			return verifyError("tokenUrl", fmt.Errorf("must be a non empty string "+
				"when the object is %v", applyTo))
		}
	case "authorizationCode":
		if o.AuthorizationUrl == "" {
			return verifyError("authorizationUrl", fmt.Errorf("must be a non empty string "+
				"when the object is %v", applyTo))
		}
		if o.TokenUrl == "" {
			return verifyError("tokenUrl", fmt.Errorf("must be a non empty string "+
				"when the object is %v", applyTo))
		}
	}
	if o.Scopes == nil {
		return verifyError("scopes", fmt.Errorf("must be a non empty object"))
	}

	if o.Extensions != nil {
		if err := validatorExtensions(o.Extensions); err != nil {
			return verifyError("extensions", err)
		}
	}
	return nil
}

type SecurityRequirement map[string][]string

type marshalField struct {
	key       string
	value     any
	omitempty bool
}

func marshalJson(list []marshalField, extensions ...map[string]any) ([]byte, error) {
	m := map[string]any{}
	if !(len(list) > 0 && list[0].key == "$ref") {
		for _, val := range extensions {
			for k, v := range val {
				m[k] = v
			}
		}
	}
	for _, v := range list {
		if v.omitempty {
			continue
		}
		m[v.key] = v.value
	}
	return packageJsonByMap(m)
}

func packageJsonByMap(m map[string]any) (buf []byte, err error) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	str := "{"
	for _, k := range keys {
		str += `"` + k + `":`
		val := reflect.ValueOf(m[k])
		if val.Kind() == reflect.Map {
			childM := map[string]any{}
			for _, key := range val.MapKeys() {
				childM[key.String()] = val.MapIndex(key).Interface()
			}
			buf, err = packageJsonByMap(childM)
		} else {
			buf, err = json.Marshal(m[k])
		}
		if err != nil {
			return nil, err
		}
		str += string(buf)
		str += ","
	}
	if len(keys) > 0 {
		str = str[:len(str)-1]
	}
	str += "}"
	return []byte(str), nil
}

func verifyError(field string, err error, isMapOrArray ...bool) error {
	errStr := err.Error()
	reg := regexp.MustCompile(`^verify (.*?) error: (.+)$`)
	paths := reg.FindStringSubmatch(errStr)
	if len(paths) == 3 {
		errStr = paths[2]
		oldField := paths[1]
		if len(isMapOrArray) > 0 && isMapOrArray[0] {
			if idx := strings.Index(oldField, "."); idx != -1 {
				oldField = "[" + oldField[:idx] + "]" + oldField[idx:]
			} else if idx = strings.Index(oldField, "["); idx != -1 {
				oldField = "[" + oldField[:idx] + "]" + oldField[idx:]
			} else {
				oldField = "[" + oldField + "]"
			}
		} else {
			oldField = "." + oldField
		}
		field += oldField
	}
	return fmt.Errorf("verify %s error: %s", field, errStr)
}

func validatorExtensions(extensions map[string]any) error {
	for k := range extensions {
		if len(k) < 2 || k[:2] != "x-" {
			return fmt.Errorf("the extended fields name must begin with 'x-'")
		}
	}
	return nil
}

func validatorRef(ref, refType string, openapi *OpenAPI) error {
	if ref == "" || ref[0] != '#' {
		return nil
	}
	if strings.TrimPrefix(ref, "#/") == "" {
		return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
	}
	if strings.Contains(ref, "{") || strings.Contains(ref, "}") {
		return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
	}
	refList := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	var api any
	api = openapi
	for i := 0; i < len(refList); i++ {
		refName := refList[i]
		// restore ref
		refName = strings.ReplaceAll(refName, "~1", "/")
		refName = strings.ReplaceAll(refName, "%7B", "{")
		refName = strings.ReplaceAll(refName, "%7D", "}")
		switch val := api.(type) {
		case *OpenAPI:
			switch refName {
			case "paths":
				api = val.Paths
			case "webhooks":
				api = val.Webhooks
			case "components":
				api = val.Components
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case *Paths: // paths
			if val.Value(refName) == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val.Value(refName)
		case *PathItem:
			switch refName {
			case "get":
				api = val.Get
			case "put":
				api = val.Put
			case "post":
				api = val.Post
			case "delete":
				api = val.Delete
			case "options":
				api = val.Options
			case "head":
				api = val.Head
			case "patch":
				api = val.Patch
			case "trace":
				api = val.Trace
			case "parameters":
				api = val.Parameters
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case *Operation:
			switch refName {
			case "parameters":
				api = val.Parameters
			case "requestBody":
				api = val.RequestBody
			case "responses":
				api = val.Responses
			case "callbacks":
				api = val.Callbacks
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case *RequestBody:
			switch refName {
			case "content":
				api = val.Content
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*MediaType:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case *MediaType:
			switch refName {
			case "schema":
				api = val.Schema
			case "examples":
				api = val.Examples
			case "encoding":
				api = val.Encoding
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case *Schema:
			switch refName {
			case "items":
				api = val.Items
			case "properties":
				api = val.Properties
			case "oneOf":
				api = val.OneOf
			case "anyOf":
				api = val.AnyOf
			case "allOf":
				api = val.AllOf
			case "not":
				api = val.Not
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*Schema:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case []*Schema:
			num, err := strconv.Atoi(refName)
			if err != nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			if val[num] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[num]
		case map[string]*Example:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case map[string]*Encoding:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case *Encoding:
			switch refName {
			case "headers":
				api = val.Headers
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*Header:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case *Header:
			switch refName {
			case "schema":
				api = val.Schema
			case "examples":
				api = val.Examples
			case "content":
				api = val.Content
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case *Responses:
			if refName == "default" {
				api = val.Default
			} else {
				if val.Value(refName) == nil {
					return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
				}
				api = val.Value(refName)
			}
		case *Response:
			switch refName {
			case "headers":
				api = val.Headers
			case "content":
				api = val.Content
			case "links":
				api = val.Links
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*Link:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case map[string]*Callback:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case *Callback:
			if val.Value(refName) == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val.Value(refName)
		case []*Parameter:
			num, err := strconv.Atoi(refName)
			if err != nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			if val[num] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[num]
		case *Parameter:
			switch refName {
			case "schema":
				api = val.Schema
			case "examples":
				api = val.Examples
			case "content":
				api = val.Content
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*PathItem: // webhooks
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case *Components: // components
			switch refName {
			case "schemas":
				api = val.Schemas
			case "responses":
				api = val.Responses
			case "parameters":
				api = val.Parameters
			case "examples":
				api = val.Examples
			case "requestBodies":
				api = val.RequestBodies
			case "headers":
				api = val.Headers
			case "securitySchemes":
				api = val.SecuritySchemes
			case "links":
				api = val.Links
			case "callbacks":
				api = val.Callbacks
			case "pathItems":
				api = val.PathItems
			default:
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
		case map[string]*Response:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case map[string]*Parameter:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case map[string]*RequestBody:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		case map[string]*SecurityScheme:
			if val[refName] == nil {
				return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
			}
			api = val[refName]
		default:
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	}
	switch refType {
	case "schema":
		if _, ok := api.(*Schema); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "response":
		if _, ok := api.(*Response); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "parameter":
		if _, ok := api.(*Parameter); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "example":
		if _, ok := api.(*Example); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "requestBody":
		if _, ok := api.(*RequestBody); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "header":
		if _, ok := api.(*Header); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "securityScheme":
		if _, ok := api.(*SecurityScheme); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "link":
		if _, ok := api.(*Link); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "callback":
		if _, ok := api.(*Callback); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "pathItem":
		if _, ok := api.(*PathItem); !ok {
			return verifyError("ref", fmt.Errorf("%q found unresolved", ref))
		}
	case "operation":
		if _, ok := api.(*Operation); !ok {
			return verifyError("operationRef", fmt.Errorf("%q found unresolved", ref))
		}
	}
	return nil
}

func handlePath(path string) (hPath string, err error) {
	var builder strings.Builder
	var symbol bool
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '{':
			if symbol {
				err = fmt.Errorf("path format error")
				return
			}
			symbol = true
			builder.WriteByte(path[i])
		case '}':
			if !symbol {
				err = fmt.Errorf("path format error")
				return
			}
			symbol = false
			builder.WriteByte(path[i])
		default:
			if symbol {
				if builder.String()[builder.Len()-1] == '{' {
					builder.WriteByte('-')
				}
			} else {
				builder.WriteByte(path[i])
			}
		}
	}
	if symbol {
		err = fmt.Errorf("path format error")
		return
	}
	hPath = builder.String()
	return
}

package goapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"

	"gopkg.in/yaml.v3"
)

type MediaTypeAnalysis interface {
	// Marshal Serialized data
	Marshal(v any) ([]byte, error)

	// Unmarshal Deserialized data
	Unmarshal([]byte, any) error

	// DefaultName swagger default name resolution
	DefaultName(name string) string

	// Info 'mediaType' is the parsing type, 'tag' is the type alias passed in as body
	Info() (mediaType MediaType, tag string)
}

func init() {
	allMediaType.setMediaTypeAnalysis(&defaultJsonAnalysis{})
	allMediaType.setMediaTypeAnalysis(&defaultXmlAnalysis{})
	allMediaType.setMediaTypeAnalysis(&defaultYamlAnalysis{})
}

// json default analysis
type defaultJsonAnalysis struct{}

func (*defaultJsonAnalysis) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (*defaultJsonAnalysis) Unmarshal(buf []byte, v any) error {
	return json.Unmarshal(buf, v)
}

func (*defaultJsonAnalysis) DefaultName(name string) string {
	return name
}

func (*defaultJsonAnalysis) Info() (mediaType MediaType, tag string) {
	return JSON, "json"
}

// xml default analysis
type defaultXmlAnalysis struct{}

func (*defaultXmlAnalysis) Marshal(v any) ([]byte, error) {
	b := new(bytes.Buffer)
	b.WriteString(xml.Header)
	body, err := xml.Marshal(v)
	if err != nil {
		return nil, err
	}
	b.Write(body)
	return b.Bytes(), nil
}

func (*defaultXmlAnalysis) Unmarshal(buf []byte, v any) error {
	return xml.Unmarshal(buf, v)
}

func (*defaultXmlAnalysis) DefaultName(name string) string {
	return name
}

func (*defaultXmlAnalysis) Info() (mediaType MediaType, tag string) {
	return XML, "xml"
}

// yaml default analysis
type defaultYamlAnalysis struct{}

func (*defaultYamlAnalysis) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (*defaultYamlAnalysis) Unmarshal(buf []byte, v any) error {
	return yaml.Unmarshal(buf, v)
}

func (*defaultYamlAnalysis) DefaultName(name string) string {
	return strings.ToLower(name)
}

func (*defaultYamlAnalysis) Info() (mediaType MediaType, tag string) {
	return YAML, "yaml"
}

package docs

import (
	"gopkg.in/yaml.v2"
	"html/template"
	"io"
	"os"
)

var (
	parseFile           = template.ParseFiles
	openFile            = os.Open
	createFile          = os.Create
	ioReadAll           = io.ReadAll
	yamlUnmarshal       = yaml.UnmarshalStrict
	executeTemplateFile = func(templateFile *template.Template, wr io.Writer, data any) error {
		return templateFile.Execute(wr, data)
	}
)

type Info struct {
	Description string `yaml:"description"`
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
}

type Servers struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
}

type SwaggerSpec struct {
	Openapi        string    `yaml:"openapi"`
	Info           Info      `yaml:"info"`
	Servers        []Servers `yaml:"servers"`
	Paths          any       `yaml:"paths"`
	Components     any       `yaml:"components"`
	SwaggerVersion string    `yaml:"x-original-swagger-version"`
}

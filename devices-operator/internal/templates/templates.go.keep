package templates

import (
	"bytes"
	"embed"
	"text/template"

	"sigs.k8s.io/yaml"
)

//go:embed *.yaml
var templates embed.FS

func LoadTemplate(name string, data any, obj interface{}) error {
	// Define the desired PersistentVolume resource
	type PersistenceConfig struct {
		Name string
	}

	// Read the content of the embedded file
	pvContent, err := templates.ReadFile("pv.yaml")
	if err != nil {
		panic("Error reading pv.yaml file")
	}

	tmpl, err := template.New("yamlTemplate").Parse(string(pvContent))
	if err != nil {
		return err
	}

	var yamlBuffer bytes.Buffer
	if err := tmpl.Execute(&yamlBuffer, data); err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlBuffer.Bytes(), obj); err != nil {
		return err
	}
	return nil
}

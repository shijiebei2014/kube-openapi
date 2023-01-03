package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/jsonreference"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/generated"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"os"
	"path/filepath"
	"strings"
)

type Property struct {
	Title string `json:"title"`
	Type  string `json:"type"`

	Const string `json:"const,omitempty"` // 常量
	Enum []interface{} `json:"enum,omitempty"` // 枚举类型
	Properties map[string]Property `json:"properties,omitempty"` // 对象类型
	Items map[string]map[string]Property `json:"items,omitempty"` // 数组类型
}

type JSONSchema struct {
	Properties map[string]Property `json:"properties"`
	Title       string        `json:"title"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Required    []interface{} `json:"required,omitempty"`
}

func main() {
	openAPIDefinitions := generated.GetOpenAPIDefinitions(func(path string) spec.Ref {
		return spec.Ref{
			Ref: jsonreference.MustCreateRef(path),
		}
	})

	for api, definition := range openAPIDefinitions {
		//if !strings.HasSuffix(api, "v1.Service") {
		//	continue
		//}
		jsonSchema := JSONSchema{
			Title: resourceName(api),
			Description: definition.Schema.Description,
			Type: "object",
		}

		jsonSchema.Properties = convert(openAPIDefinitions, definition.Schema.SchemaProps.Properties, api)

		err := writeToFile(api, jsonSchema)
		if err != nil {
			fmt.Printf("%s resolve err: %s\n", api, err.Error())
		}
	}
}

func writeToFile(api string, schema JSONSchema) error {
	fileName, err := filepath.Abs(fmt.Sprintf("doc/schema/%s.json", resourceName(api)))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(schema)
}

func convert(openAPIDefinitions map[string]common.OpenAPIDefinition, props map[string]spec.Schema, api string) map[string]Property {
	properties := make(map[string]Property)
	for field, prop := range props {
		types := prop.Type
		if len(types) == 0 {
			ref := prop.SchemaProps.Ref.String()
			definition := openAPIDefinitions[ref]

			properties[field] = Property{
				Title: field,
				Type:  "object",
				Properties: convert(openAPIDefinitions, definition.Schema.SchemaProps.Properties, ref),
			}
			continue
		}

		if types[0] == "array" {
			//if strings.HasSuffix(field, "ports") {
			//	fmt.Println("111")
			//}
			ref := prop.SchemaProps.Items.Schema.SchemaProps.Ref.Ref.String()
			definition := openAPIDefinitions[ref]
			properties[field] = Property{
				Title: field,
				Type:  "array",
				Items: map[string]map[string]Property{
					"properties": convert(openAPIDefinitions, definition.Schema.SchemaProps.Properties, ref),
				},
			}
			continue
		}

		property := Property{
			Title: field,
			Type:  types[0],
			Enum:  prop.Enum,
		}

		switch field {
		case "kind":
			_, property.Const = apiVersionAndKind(api)
		case "apiVersion":
			property.Const, _ = apiVersionAndKind(api)
		}

		properties[field] = property
	}
	return properties
}

func resourceName(api string) string {
	apiVersion, kind := apiVersionAndKind(api)
	return fmt.Sprintf("%s_%s", strings.ReplaceAll(apiVersion, "/", "_"), kind)
}

func apiVersionAndKind(api string) (apiVersion, kind string) {
	strs := strings.Split(api, "/")
	if len(strs) < 1 {
		return
	}
	last := strs[len(strs) - 1]
	kind = strings.Split(last, ".")[1]
	apiVersion = fmt.Sprintf("%s/%s",strs[len(strs) - 2], strings.Split(last, ".")[0])
	return
}

/*
	struct {
		Shard struct {
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"shard"`
		ConvertShard struct {
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"convertShard"`
		MinDateYear struct {
			Title   string `json:"title"`
			Type    string `json:"type"`
			Minimum int    `json:"minimum"`
			Maximum int    `json:"maximum"`
		} `json:"minDateYear"`
		MinDateMonth struct {
			Title   string `json:"title"`
			Type    string `json:"type"`
			Minimum int    `json:"minimum"`
			Maximum int    `json:"maximum"`
		} `json:"minDateMonth"`
		Indexes struct {
			Items struct {
				Properties struct {
					Key struct {
						Properties struct {
						} `json:"properties"`
						Title string `json:"title"`
						Type  string `json:"type"`
					} `json:"key"`
					Unique struct {
						Title string `json:"title"`
						Type  string `json:"type"`
					} `json:"unique"`
				} `json:"properties"`
				Title string `json:"title"`
				Type  string `json:"type"`
			} `json:"items"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"indexes"`
*/
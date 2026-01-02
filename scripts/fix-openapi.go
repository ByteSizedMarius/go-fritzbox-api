//go:build ignore

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Helper to find a key in a mapping node and return its value node
func findKey(mapping *yaml.Node, key string) *yaml.Node {
	if mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// Helper to set or replace a key's value in a mapping node
func setKey(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	// Key not found, append it
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		value,
	)
}

// Helper to create a $ref node
func createRef(refPath string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "$ref"},
			{Kind: yaml.ScalarNode, Value: refPath},
		},
	}
}

// fixButtonConfig replaces IF_button_config's anyOf/discriminator pattern with IF_button_config_units
// This allows oapi-codegen to generate a proper typed struct instead of map[string]interface{}
func fixButtonConfig(schemas *yaml.Node) {
	if schemas.Kind != yaml.MappingNode {
		return
	}

	// Find IF_button_config
	buttonConfig := findKey(schemas, "IF_button_config")
	if buttonConfig == nil {
		return
	}

	// Replace entire schema with just $ref to IF_button_config_units (the superset with all fields)
	setKey(schemas, "IF_button_config", createRef("#/components/schemas/IF_button_config_units"))
}

// extractInlineSchemas extracts inline schemas from IF_thermostat_config and adds them as named schemas
func extractInlineSchemas(schemas *yaml.Node) {
	if schemas.Kind != yaml.MappingNode {
		return
	}

	// Find IF_thermostat_config
	thermostatConfig := findKey(schemas, "IF_thermostat_config")
	if thermostatConfig == nil {
		return
	}

	// IF_thermostat_config uses allOf, properties are in the second element
	allOf := findKey(thermostatConfig, "allOf")
	if allOf == nil || allOf.Kind != yaml.SequenceNode || len(allOf.Content) < 2 {
		return
	}

	// The second element of allOf contains the properties
	secondSchema := allOf.Content[1]
	properties := findKey(secondSchema, "properties")
	if properties == nil {
		return
	}

	// Extract temperatureOffset
	extractTemperatureOffset(schemas, properties)

	// Extract windowOpenMode
	extractWindowOpenMode(schemas, properties)
}

// extractTemperatureOffset extracts the inline temperatureOffset schema
func extractTemperatureOffset(schemas, properties *yaml.Node) {
	// Find temperatureOffset in properties
	for i := 0; i < len(properties.Content); i += 2 {
		if properties.Content[i].Value == "temperatureOffset" {
			inlineSchema := properties.Content[i+1]

			// Add the schema to components/schemas as helper_temperatureOffset
			schemas.Content = append(schemas.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "helper_temperatureOffset"},
				inlineSchema,
			)

			// Replace inline with $ref
			properties.Content[i+1] = createRef("#/components/schemas/helper_temperatureOffset")
			return
		}
	}
}

// extractWindowOpenMode extracts the inline windowOpenMode allOf extension
func extractWindowOpenMode(schemas, properties *yaml.Node) {
	// Find windowOpenMode in properties
	for i := 0; i < len(properties.Content); i += 2 {
		if properties.Content[i].Value == "windowOpenMode" {
			inlineSchema := properties.Content[i+1]

			// Add the schema to components/schemas as helper_windowOpenModeConfig
			schemas.Content = append(schemas.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "helper_windowOpenModeConfig"},
				inlineSchema,
			)

			// Replace inline with $ref
			properties.Content[i+1] = createRef("#/components/schemas/helper_windowOpenModeConfig")
			return
		}
	}
}

// Recursively remove readOnly from all schemas to avoid allOf merge conflicts
func removeReadOnly(node *yaml.Node) {
	if node == nil {
		return
	}

	if node.Kind == yaml.MappingNode {
		// Find and remove readOnly keys
		newContent := make([]*yaml.Node, 0, len(node.Content))
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Value == "readOnly" {
				// Skip this key-value pair
				continue
			}
			newContent = append(newContent, node.Content[i], node.Content[i+1])
			// Recurse into value
			removeReadOnly(node.Content[i+1])
		}
		node.Content = newContent
	} else if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			removeReadOnly(child)
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run fix-openapi.go <input.yaml> <output.yaml>")
		os.Exit(1)
	}

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(input, &doc); err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Process components/schemas section
	if doc.Content != nil && len(doc.Content) > 0 {
		root := doc.Content[0]
		if root.Kind == yaml.MappingNode {
			for i := 0; i < len(root.Content); i += 2 {
				if root.Content[i].Value == "components" {
					components := root.Content[i+1]
					if components.Kind == yaml.MappingNode {
						for j := 0; j < len(components.Content); j += 2 {
							if components.Content[j].Value == "schemas" {
								schemas := components.Content[j+1]
								removeReadOnly(schemas)
								extractInlineSchemas(schemas)
								fixButtonConfig(schemas)
							}
						}
					}
				}
			}
		}
	}

	output, err := yaml.Marshal(&doc)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(os.Args[2], output, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fixed YAML written to %s\n", os.Args[2])
}

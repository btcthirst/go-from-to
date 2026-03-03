// Package mappings loads and provides column mappings for different bank statement formats.
package mappings

import (
	"fmt"
	"log"
	"os"

	"bank-analyzer/assets"

	"gopkg.in/yaml.v3"
)

// ParserMapping holds column mappings for a single parser.
// Key is a logical field name (date, amount, ...), value is a list of possible column names.
type ParserMapping map[string][]string

// MappingsConfig holds column mappings for all parsers.
type MappingsConfig map[string]ParserMapping

// Get returns the list of column name candidates for a given field.
// Returns nil if the field is not found — callers treat this as index -1.
func (m ParserMapping) Get(field string) []string {
	if candidates, ok := m[field]; ok {
		return candidates
	}
	return nil
}

// Load reads column mappings from a YAML file.
// On error — falls back to the embedded file so the app never crashes.
//
// Merge strategy: the embedded file is the base. The file on disk overrides
// only the parsers it defines. Parsers absent from the file keep embedded defaults.
func Load(path string) MappingsConfig {
	embedded := mustParseMappingsConfig(assets.ColumnMappingsYAML)

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[mappings] warn: could not read %s: %v, using embedded file", path, err)
		return embedded
	}

	cfg, err := parseMappingsConfig(data)
	if err != nil {
		log.Printf("[mappings] warn: could not parse %s: %v, using embedded file", path, err)
		return embedded
	}

	merged := make(MappingsConfig, len(embedded))
	for k, v := range embedded {
		merged[k] = v
	}
	for k, v := range cfg {
		merged[k] = v
	}

	return merged
}

// MustGetParser returns the mapping for a specific parser key.
// Panics if the key is not found — a missing key means a developer registered
// a parser without adding its mapping to column_mappings.yaml, which is a
// programming error that must be caught at startup, not silently ignored.
//
// Follows the standard Go "Must" convention (regexp.MustCompile, template.Must).
func (c MappingsConfig) MustGetParser(parserKey string) ParserMapping {
	m, ok := c[parserKey]
	if !ok {
		panic(fmt.Sprintf("mapping for parser %q not found — add the key to column_mappings.yaml", parserKey))
	}
	return m
}

func parseMappingsConfig(data []byte) (MappingsConfig, error) {
	var cfg MappingsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// mustParseMappingsConfig parses the embedded YAML and panics on error.
// The embedded file is part of the binary and must always be valid.
func mustParseMappingsConfig(data []byte) MappingsConfig {
	cfg, err := parseMappingsConfig(data)
	if err != nil {
		panic("embedded column_mappings.yaml is corrupted: " + err.Error())
	}
	return cfg
}

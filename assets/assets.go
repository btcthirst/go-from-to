// Package assets містить вбудовані файли конфігурації за замовчуванням.
// Файли з цього пакету використовуються як fallback коли зовнішні
// конфігураційні файли відсутні або пошкоджені.
package assets

import _ "embed"

//go:embed categories.yaml
var CategoriesYAML []byte

//go:embed column_mappings.yaml
var ColumnMappingsYAML []byte

//go:embed report_311.yaml
var Report311YAML []byte

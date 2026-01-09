package bank_parsers

import (
	"billing-note/internal/pdf"
)

// RegisterAllParsers registers all bank parsers to the registry
func RegisterAllParsers(registry *pdf.ParserRegistry) {
	registry.RegisterParser(NewCathayParser())
	registry.RegisterParser(NewTaishinParser())
	registry.RegisterParser(NewFubonParser())
}

// NewRegistryWithAllParsers creates a new registry with all parsers registered
func NewRegistryWithAllParsers() *pdf.ParserRegistry {
	registry := pdf.NewParserRegistry()
	RegisterAllParsers(registry)
	return registry
}

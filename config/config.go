package config

// AppConfig Global Application Configuration
var AppConfig = struct {
	// EnableImport controls whether the "Import Existing Resume" feature is available.
	// Set to true to show the import form on the homepage and enable the /import endpoint.
	// Set to false to hide it.
	EnableImport            bool
	EnableAIAssistant       bool
	EnableTemplateSelection bool
}{
	EnableImport:            true,
	EnableAIAssistant:       true,
	EnableTemplateSelection: true,
}

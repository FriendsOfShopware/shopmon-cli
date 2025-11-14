package deployment

import (
	"encoding/json"
	"os"
)

// ComposerReader interface for reading composer.json files (allows mocking in tests)
type ComposerReader interface {
	ReadComposerData(filepath string) (map[string]interface{}, error)
}

// DefaultComposerReader implements ComposerReader
type DefaultComposerReader struct{}

// NewDefaultComposerReader creates a new DefaultComposerReader
func NewDefaultComposerReader() *DefaultComposerReader {
	return &DefaultComposerReader{}
}

// ReadComposerData reads and parses composer.json file
func (r *DefaultComposerReader) ReadComposerData(filepath string) (map[string]interface{}, error) {
	composerData := make(map[string]interface{})

	composerFile, err := os.ReadFile(filepath)
	if err != nil {
		// Return empty map if file doesn't exist
		if os.IsNotExist(err) {
			return composerData, nil
		}
		return nil, err
	}

	var composer ComposerJSON
	if err := json.Unmarshal(composerFile, &composer); err != nil {
		return nil, err
	}

	// Convert require map to the format we need
	for pkg, version := range composer.Require {
		composerData[pkg] = version
	}

	return composerData, nil
}
package deployment

import (
	"encoding/json"
	"os"
)

// ReadComposerData reads and parses composer.json file
func ReadComposerData(filepath string) (map[string]interface{}, error) {
	composerData := make(map[string]interface{})

	composerFile, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return composerData, nil
		}
		return nil, err
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(composerFile, &composer); err != nil {
		return nil, err
	}

	for pkg, version := range composer.Require {
		composerData[pkg] = version
	}

	return composerData, nil
}

package deployment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultComposerReader_ReadComposerData(t *testing.T) {
	reader := NewDefaultComposerReader()

	t.Run("valid composer.json", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		composerFile := filepath.Join(tmpDir, "composer.json")

		content := `{
			"require": {
				"php": ">=8.1",
				"shopware/core": "6.5.0.0",
				"shopware/administration": "6.5.0.0"
			}
		}`

		err := os.WriteFile(composerFile, []byte(content), 0644)
		require.NoError(t, err)

		data, err := reader.ReadComposerData(composerFile)
		require.NoError(t, err)
		require.NotNil(t, data)

		assert.Equal(t, ">=8.1", data["php"])
		assert.Equal(t, "6.5.0.0", data["shopware/core"])
		assert.Equal(t, "6.5.0.0", data["shopware/administration"])
		assert.Len(t, data, 3)
	})

	t.Run("non-existent file", func(t *testing.T) {
		data, err := reader.ReadComposerData("/nonexistent/composer.json")
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		composerFile := filepath.Join(tmpDir, "composer.json")

		err := os.WriteFile(composerFile, []byte("invalid json"), 0644)
		require.NoError(t, err)

		data, err := reader.ReadComposerData(composerFile)
		assert.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("empty require section", func(t *testing.T) {
		tmpDir := t.TempDir()
		composerFile := filepath.Join(tmpDir, "composer.json")

		content := `{
			"name": "test/package",
			"require": {}
		}`

		err := os.WriteFile(composerFile, []byte(content), 0644)
		require.NoError(t, err)

		data, err := reader.ReadComposerData(composerFile)
		require.NoError(t, err)
		require.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("no require section", func(t *testing.T) {
		tmpDir := t.TempDir()
		composerFile := filepath.Join(tmpDir, "composer.json")

		content := `{
			"name": "test/package",
			"description": "A test package"
		}`

		err := os.WriteFile(composerFile, []byte(content), 0644)
		require.NoError(t, err)

		data, err := reader.ReadComposerData(composerFile)
		require.NoError(t, err)
		require.NotNil(t, data)
		assert.Empty(t, data)
	})

	t.Run("complex composer.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		composerFile := filepath.Join(tmpDir, "composer.json")

		content := `{
			"name": "shopware/production",
			"description": "Shopware 6 production template",
			"type": "project",
			"license": "MIT",
			"require": {
				"php": ">=8.1",
				"ext-dom": "*",
				"ext-json": "*",
				"shopware/core": "6.5.0.0",
				"shopware/administration": "6.5.0.0",
				"shopware/storefront": "6.5.0.0",
				"symfony/flex": "^2.0"
			},
			"require-dev": {
				"phpunit/phpunit": "^9.5"
			}
		}`

		err := os.WriteFile(composerFile, []byte(content), 0644)
		require.NoError(t, err)

		data, err := reader.ReadComposerData(composerFile)
		require.NoError(t, err)
		require.NotNil(t, data)

		// Should only include "require" section, not "require-dev"
		assert.Equal(t, ">=8.1", data["php"])
		assert.Equal(t, "*", data["ext-dom"])
		assert.Equal(t, "*", data["ext-json"])
		assert.Equal(t, "6.5.0.0", data["shopware/core"])
		assert.Equal(t, "^2.0", data["symfony/flex"])
		assert.Len(t, data, 7)

		// require-dev should not be included
		_, hasPhpUnit := data["phpunit/phpunit"]
		assert.False(t, hasPhpUnit)
	})
}

func TestComposerJSON_Structure(t *testing.T) {
	composer := ComposerJSON{
		Require: map[string]string{
			"php":          ">=8.1",
			"package/name": "1.0.0",
		},
	}

	assert.Len(t, composer.Require, 2)
	assert.Equal(t, ">=8.1", composer.Require["php"])
	assert.Equal(t, "1.0.0", composer.Require["package/name"])
}
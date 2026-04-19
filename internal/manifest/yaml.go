package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"

	"go.yaml.in/yaml/v3"
)

func LoadYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func WriteYAML(path string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, data)
}

func WriteYAMLExclusive(path string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return fsutil.WriteFileExclusive(path, data)
}

func WriteYAMLIfMissing(path string, value any) error {
	if fsutil.Exists(path) {
		return nil
	}
	return WriteYAML(path, value)
}

func IsMissing(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

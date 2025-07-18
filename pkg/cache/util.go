package cache

import (
	"errors"
	"os"
)

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

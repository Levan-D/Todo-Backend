package utils

import (
	"os"
)

func CreateDirectoryIfNotExists(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

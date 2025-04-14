package cli

import (
	"errors"
	flag "github.com/spf13/pflag"
	"os"
)

type Command struct {
	Flags *flag.FlagSet
	Run   func(flags *flag.FlagSet)
}

// ValidatorFile only allows file paths
func ValidatorFile(val string) error {
	s, err := os.Stat(val)
	if os.IsNotExist(err) {
		return errors.New("Path has to represent an existing file")
	}
	if s.IsDir() {
		return errors.New("Path has to represent a file")
	}
	return nil
}

// ValidatorDir only allows directory paths
func ValidatorDir(val string) error {
	s, err := os.Stat(val)
	if os.IsNotExist(err) {
		return errors.New("Path has to represent an existing dir")
	}
	if !s.IsDir() {
		return errors.New("Path has to represent a dir")
	}
	return nil
}

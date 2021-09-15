package cli

import (
	"errors"
	"flag"
	"os"
)

var SpinnerCharSet = []string{"[       ]", "[=      ]", "[==     ]", "[===    ]", "[====   ]", "[ ====  ]", "[  ==== ]", "[   ====]", "[    ===]", "[     ==]", "[      =]"}

type Command struct {
	Flags *flag.FlagSet
	Run   func(flags *flag.FlagSet)
}

// ValidatorFile only allows file paths
func ValidatorFile(val interface{}) error {
	p, ok := val.(string)
	if !ok {
		panic("cli: validator only applies to string values")
	}
	s, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("Path has to represent an existing file")
	}
	if s.IsDir() {
		return errors.New("Path has to represent a file")
	}
	return nil
}

// ValidatorDir only allows directory paths
func ValidatorDir(val interface{}) error {
	p, ok := val.(string)
	if !ok {
		panic("cli: validator only applies to string values")
	}
	s, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("Path has to represent an existing dir")
	}
	if !s.IsDir() {
		return errors.New("Path has to represent a dir")
	}
	return nil
}

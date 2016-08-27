package factory

import (
	"fmt"
	"os"
)

type DrvInterface interface {
	New() (DrvInterface, error)
	Get(file string) ([]byte, error)
	Save(file string) (string, error)
	ReadStream(path string, offset uint64) (*os.File, error)
	Delete(file string) error
}

var Drivers = make(map[string]DrvInterface)

func Register(name string, driver DrvInterface) error {
	if _, existed := Drivers[name]; existed {
		return fmt.Errorf("%v has already been registered", name)
	}

	Drivers[name] = driver

	return nil
}

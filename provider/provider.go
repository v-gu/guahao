package provider

import (
	"fmt"

	config "github.com/v-gu/guahao/config"
	driver "github.com/v-gu/guahao/provider/driver"
)

var drivers = make(map[string]driver.Driver)

// Register makes a booking provider availiable by provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver driver.Driver) {
	if driver == nil {
		panic("provider: Register driver is null")
	}
	for _, n := range config.All.Disabled {
		if n == name {
			// skip disabled driver
			return
		}
	}
	if _, dup := drivers[name]; dup {
		panic(fmt.Sprintf("provider: Register called twice for driver [%v]", name))
	}
	drivers[name] = driver

	// fill config
	err := config.All.UnmarshalConfig(name, driver)
	if err != nil {
		panic(fmt.Sprintf("can't unmarshal config for driver: [%v]: %s", name, err))
	}
}

//
func Login() error {
	err := drivers["zjol"].Login()
	if err != nil {
		return err
	}
	return nil
}

func Book() error {
	err := drivers["zjol"].Book()
	if err != nil {
		return err
	}
	return nil
}

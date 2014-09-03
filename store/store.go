package store

import (
	"errors"
	"os"
	"path/filepath"

	toml "github.com/BurntSushi/toml"

	config "github.com/v-gu/guahao/config"
	log "github.com/v-gu/guahao/log"
)

var (
	Store Storage = Storage{NamedLogger: log.NamedLogger{"store"}}
)

func init() {
	Store.path = config.All.StorePath
}

type Storage struct {
	path string
	log.NamedLogger
}

// Marshal a interface value to a Toml format disk file, the 'elem's
// should be relative path name element(s) for the storage file.
func (s *Storage) Marshal(v interface{}, elem ...string) (err error) {
	if len(elem) == 0 {
		return errors.New("no name component(s) provided")
	}
	elems := []string{s.path}
	elems = append(elems, elem...)
	fDir := filepath.Join(elems[:len(elems)-1]...)
	fFile := filepath.Join(elems...)
	s.Debugf(log.DEBUG_CONFIG, "mkdir: [%v]\n", fDir)
	err = os.MkdirAll(fDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}
	s.Debugf(log.DEBUG_CONFIG, "open file for write: [%v]\n", fFile)
	file, err := os.Create(fFile)
	if err != nil {
		return
	}
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(v)
	return
}

// Unmarshal a interface value from a Toml format disk file, the 'names'
// should be relative path name element(s) for the storage directory.
func (s *Storage) Unmarshal(v interface{}, elem ...string) (err error) {
	if len(elem) == 0 {
		return errors.New("no name component(s) provided")
	}
	elems := []string{s.path}
	elems = append(elems, elem...)
	fPath := filepath.Join(elems...)
	s.Debugf(log.DEBUG_CONFIG, "open file for read: [%v]\n", fPath)
	_, err = toml.DecodeFile(fPath, v)
	return
}

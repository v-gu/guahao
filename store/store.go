package store

import (
	"errors"
	"os"
	"path/filepath"

	toml "github.com/BurntSushi/toml"
	glog "github.com/golang/glog"

	config "github.com/v-gu/guahao/config"
)

var (
	Store Storage
)

func init() {
	Store.path = config.All.StorePath
}

type Storage struct {
	path string
}

// Marshal a interface value to a Toml format disk file, the 'names'
// should be relative path name element(s) for the storage directory.
func (s *Storage) Marshal(v interface{}, elem ...string) (err error) {
	if len(elem) == 0 {
		return errors.New("storage: no name component(s) provided")
	}
	elems := []string{s.path}
	elems = append(elems, elem...)
	fDir := filepath.Join(elems[:len(elems)-1]...)
	fFile := filepath.Join(elems...)
	if glog.V(config.LOG_CONFIG) {
		glog.Infof("storage: mkdir: [%v]\n", fDir)
	}
	err = os.MkdirAll(fDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}
	if glog.V(config.LOG_CONFIG) {
		glog.Infof("storage: open file for write: [%v]\n", fFile)
	}
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
		return errors.New("storage: no name component(s) provided")
	}
	elems := []string{s.path}
	elems = append(elems, elem...)
	fPath := filepath.Join(elems...)
	if glog.V(config.LOG_CONFIG) {
		glog.Infof("storage: open file for read: [%v]\n", fPath)
	}
	_, err = toml.DecodeFile(fPath, v)
	return
}

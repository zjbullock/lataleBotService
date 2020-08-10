package server

import (
	"fmt"
	"github.com/juju/loggo"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var schema = ""
var walk = filepath.Walk

type server struct {
	log loggo.Logger
}

type Server interface {
	GetSchema(root string, l loggo.Logger) (string, error)
}

func NewServer(l loggo.Logger) Server {
	return &server{
		log: l,
	}
}
func (s *server) GetSchema(root string, l loggo.Logger) (string, error) {
	err := walk(root, s.checkFile)
	l.Debugf(fmt.Sprintf("Schema: \n %v", schema))
	return schema, err
}

func (s *server) checkFile(path string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".graphql") {
		s.log.Debugf(filepath.Abs(""))
		abs, _ := filepath.Abs(path)
		content, err := ioutil.ReadFile(abs)
		if err != nil {
			return err
		}
		s.log.Debugf(path)
		schema += strings.TrimSpace(string(content)) + "\n"
	} else {
		s.log.Debugf("Path: %v, is not graphql file, possible directory", path)
	}
	return nil
}

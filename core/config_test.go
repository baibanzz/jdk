package core

import (
	"log"
	"testing"

	"github.com/baibanzz/jdk/model"
)

func TestConfig(t *testing.T) {
	var s struct {
		Mysql model.Mysql
	}
	config, err := NewConfig("./config.yaml", &s)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(config, s)
}

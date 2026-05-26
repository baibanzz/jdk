package core

import (
	"github.com/baibanzz/jdk/core/internal/etcd"
	"github.com/baibanzz/jdk/model"
)

type Etcd struct {
	*etcd.Etcd
}

func NewEtcd(d model.Etcd) (*Etcd, error) {
	e, err := etcd.New(d)
	return &Etcd{e}, err
}

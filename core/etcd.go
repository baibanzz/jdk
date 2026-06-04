package core

import (
	"github.com/baibanzz/jdk/core/internal/etcd"
	"github.com/baibanzz/jdk/model"
)

type Etcd = etcd.Etcd

func NewEtcd(d model.Etcd) (*Etcd, error) {
	return etcd.New(d)
}

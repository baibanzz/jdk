package core

import (
	"github.com/baibanzz/jdk/core/internal/etcd"
	"github.com/baibanzz/jdk/model"
)

func NewEtcd(d model.Etcd) (*etcd.Etcd, error) {
	return etcd.New(d)
}

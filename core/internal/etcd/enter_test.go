package etcd

import (
	"testing"
	"time"

	"github.com/baibanzz/jdk/model"
)

var etcdData = model.Etcd{
	Host: []string{"127.0.0.1:2379"},
}

func TestEtcd_AutoPut(t *testing.T) {
	etcd, err := New(etcdData)
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	etcd.AutoPut("demo", []byte("auto"))
	time.Sleep(1 * time.Second)
	_, err = etcd.Get("demo")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Second)
	_, err = etcd.Get("demo")
	if err != nil {
		t.Fatal(err)
	}
}

func TestEtcd_PutGetDelete(t *testing.T) {
	etcd, err := New(etcdData)
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	if err := etcd.Put("demo20", []byte("111"), 180); err != nil {
		t.Fatal(err)
	}
	if err := etcd.Put("demo0", []byte("222"), 0); err != nil {
		t.Fatal(err)
	}
	list, err := etcd.GetList("")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(list)
	}
	get, err := etcd.Get("demo0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(get)
	}
	err = etcd.Delete("demo0")
	if err != nil {
		t.Fatal(err)
	}
}

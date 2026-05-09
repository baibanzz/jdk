package utils

import "testing"

func TestNewUUID(t *testing.T) {
	newUUID := NewUUID()
	t.Log(newUUID)
}

func TestStructAndMap(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	User1 := User{
		Name: "john",
		Age:  18,
	}
	toMap := structToMap(User1)
	t.Log(toMap)
	toStruct, err := mapToStruct(toMap, User{})
	if err != nil {
		t.Error(err)
	}
	t.Log(toStruct)
	if toStruct != User1 {
		t.Error("值不相同")
	}

}

func TestTo(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	User1 := User{
		Name: "john",
		Age:  18,
	}
	tomap := structToMap(User1)
	t.Log(tomap)
	json, _ := ToJson(User1)
	t.Log(string(json))

	toMap, err := mapToStruct(tomap, User{})
	if err != nil {
		t.Error(err)
	}
	t.Log(toMap)
	toStruct, err := ToStruct(json, User{})
	if err != nil {
		t.Error(err)
	}
	t.Log(toStruct)
}

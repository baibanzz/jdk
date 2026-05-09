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
	toMap := StructToMap(User1)
	t.Log(toMap)
	toStruct, err := MapToStruct(toMap, User{})
	if err != nil {
		t.Error(err)
	}
	t.Log(toStruct)
	if toStruct != User1 {
		t.Error("值不相同")
	}

}

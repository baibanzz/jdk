package jwt

import "testing"

func TestJWT(t *testing.T) {
	type demo struct {
		Name string
		Age  int
	}
	d := demo{
		Name: "haha",
		Age:  12,
	}

	ls := TokenClaims[demo]{
		Claims: d,
	}
	token, err := ls.SignToken("12345678")
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	ls.Claims = demo{
		Name: "",
		Age:  0,
	}
	err = ls.ParseToken("12345678", token)
	if err != nil {
		t.Error(err)
	}
	t.Log(ls)
}

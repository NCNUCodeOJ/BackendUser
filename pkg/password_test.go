package pkg

import "testing"

func Test(t *testing.T) {
	d, err := Encrypt("asdasd")
	if err != nil {
		t.Fail()
	}
	if Compare(d, "asdasd") != nil {
		t.Fail()
	}
}

package tests

import (
	"github.com/Gamma169/go-server-helpers/db"
	"strings"
	"testing"
)

/*********************************************
 * Tests
 * *******************************************/

// TODO: look into using this library?  Unfortunately it's old and possibly stale
// https://github.com/erikstmartin/go-testdb
func TestInitDB(t *testing.T) {
	t.Skip("TODO")
}

func TestCheckDBConnection(t *testing.T) {
	t.Skip("TODO")
}

func TestValidateDBConnOrPanic(t *testing.T) {
	t.Skip("TODO")
}

func TestAssignArrayPropertyFromString(t *testing.T) {

	type myStruct struct {
		MyArr []string
	}

	delimiter := "::"

	expectedVal := []string{"foo", "bar", "baz", "buz"}
	inputStr := strings.Join(expectedVal, delimiter)

	inputStruct := myStruct{}

	err := db.AssignArrayPropertyFromString(&inputStruct, "MyArr", inputStr, delimiter)

	ok(t, err)
	assert(t, inputStruct.MyArr != nil, "Input struct field is non nil")
	equals(t, len(expectedVal), len(inputStruct.MyArr))
	for i, v := range expectedVal {
		equals(t, v, inputStruct.MyArr[i])
	}
}

func TestCheckStructFieldsForInjection(t *testing.T) {

	type myStruct struct {
		MyArr  []string
		MyInt  int
		MyVal  interface{}
		MyStr1 string
		MyStr2 string
		MyStr3 string
	}

	s := myStruct{
		MyArr:  []string{"qe", "asd", "zxc"},
		MyInt:  5,
		MyVal:  8,
		MyStr1: "someStr",
		MyStr2: "another str",
	}

	err := db.CheckStructFieldsForInjection(s)
	ok(t, err)

	s.MyStr2 = "some semi;colon str"
	err = db.CheckStructFieldsForInjection(s)
	assert(t, err != nil, "should return error if a field has a semicolon")
}

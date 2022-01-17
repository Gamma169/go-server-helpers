package tests

import (
	"errors"
	"github.com/Gamma169/go-server-helpers/db"
	"math/rand"
	"strings"
	"testing"
)

/*********************************************
 * Tests
 * *******************************************/

func TestCheckAndRetry(t *testing.T) {

	testCases := []struct {
		maxTries          int
		errorAfterTimes   int
		recoverAfterTimes int
		shouldPass        bool
	}{
		{rand.Intn(50), 80, 100, true},
		{rand.Intn(50), 80, 100, true},
		{rand.Intn(50), 80, 100, true},
		{rand.Intn(50), 80, 100, true},
		{rand.Intn(20), 21, 40, true},
		{rand.Intn(20), 0, 100, false},
		{rand.Intn(20), 0, 100, false},
		{rand.Intn(20), 0, 100, false},
		{50, 0, 19, true},
		{20, 0, rand.Intn(20), true},
		{70, 0, rand.Intn(50), true},
	}

	for _, tc := range testCases {
		tries := 0
		var errStr string
		checkerFunc := func() error {
			tries++
			if tries > tc.errorAfterTimes && tries < tc.recoverAfterTimes {
				errStr = randString(50)
				return errors.New(errStr)
			}
			return nil
		}

		err := db.CheckAndRetry(checkerFunc, tc.maxTries, 0, false)
		if tc.shouldPass {
			ok(t, err)
		} else {
			assert(t, err != nil, "Should return error,  %d maxTries, %d errorAfterTimes, %d recoverAfterTimes", tc.maxTries, tc.errorAfterTimes, tc.recoverAfterTimes)
			assert(t, err.Error() == errStr, "Should return last error gotten")
		}
	}

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

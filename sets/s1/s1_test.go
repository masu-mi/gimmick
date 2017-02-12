package s1

import "testing"

func TestRotationNumber(t *testing.T) {
	type testCase struct {
		input    []uint64
		expected int
	}
	for _, test := range []testCase{
		testCase{expected: 0, input: []uint64{0}},
		testCase{expected: 0, input: []uint64{1}},
		testCase{expected: 0, input: []uint64{1, 1}},
		testCase{expected: 1, input: []uint64{1, 2}},
		testCase{expected: 1, input: []uint64{1, 2, 3}},
		testCase{expected: 1, input: []uint64{1, 2, 1}},
		testCase{expected: 1, input: []uint64{1, 0}},
		testCase{expected: 2, input: []uint64{1, 3, 2}},
		testCase{expected: 2, input: []uint64{1, 3, 2, 1}},
	} {
		if act := RotationNumber(test.input...); act != test.expected {
			t.Errorf("RotationNumber(%#v); expected: %d, act: %d\n", test.input, test.expected, act)
		}
	}
}

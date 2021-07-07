package fpgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromArrayMapReduce(t *testing.T) {
	var s *StreamDef
	var tempString string

	s = StreamFromArray([]MaybeDef{Just("1"), Just("2"), Just("3"), Just("4")})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)
	s = s.Map(func(index int) interface{} {
		var val = Just(s.Get(index)).ToMaybe().ToString()
		var result interface{} = "v" + val
		return result
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "v1v2v3v4", tempString)
	tempString = ""

	s = StreamFromArray([]string{"1", "2", "3", "4"})
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)
	s = s.Map(func(index int) interface{} {
		var val = Just(s.Get(index)).ToString()
		var result interface{} = "v" + val
		return result
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "v1v2v3v4", tempString)

	s = StreamFromArray([]int{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)
	s = s.Map(func(index int) interface{} {
		var val, _ = Just(s.Get(index)).ToInt()
		var result interface{} = val * val
		return result
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "14916", tempString)

	s = StreamFromArray([]float32{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)
	s = s.Map(func(index int) interface{} {
		var val, _ = Just(s.Get(index)).ToFloat32()
		var result interface{} = val * val
		return result
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "14916", tempString)

	s = StreamFromArray([]float64{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = s.Map(func(index int) interface{} {
		var val, _ = Just(s.Get(index)).ToFloat64()
		var result interface{} = val * val
		return result
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "14916", tempString)

	s = StreamFromArray([]bool{true, false, true, false})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "truefalsetruefalse", tempString)

	s = StreamFromArray([]byte{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = StreamFromArray([]int8{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = StreamFromArray([]int16{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = StreamFromArray([]int32{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = StreamFromArray([]int64{1, 2, 3, 4})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)
}

func TestFilter(t *testing.T) {
	var s *StreamDef
	var tempString string

	s = StreamFromArray([]interface{}{}).Append(1).Extend(StreamFromArray([]interface{}{2, 3, 4})).Extend(StreamFromArray([]interface{}{nil})).Extend(nil)
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234<nil>", tempString)
	s = s.Distinct()
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "1234", tempString)

	s = s.Filter(func(index int) bool {
		var val, err = Just(s.Get(index)).ToInt()

		return err == nil && val > 1 && val < 4
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "23", tempString)
}

func TestSort(t *testing.T) {
	var s *StreamDef
	var tempString string

	s = StreamFromArray([]int{11}).Extend(StreamFromArray([]int{2, 3, 4, 5})).Remove(4)
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "11234", tempString)

	s = s.Sort(func(i, j int) bool {
		var vali, _ = Just(s.Get(i)).ToInt()
		var valj, _ = Just(s.Get(j)).ToInt()
		return vali < valj
	})
	tempString = ""
	for _, v := range s.ToArray() {
		tempString += Just(v).ToMaybe().ToString()
	}
	assert.Equal(t, "23411", tempString)
}

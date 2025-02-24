package fpgo

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
)

/**
Special thanks
* fp functions(Dedupe/Difference/Distinct/IsDistinct/DropEq/Drop/DropLast/DropWhile/IsEqual/IsEqualMap/Every/Exists/Intersection/Keys/Values/Max/Min/MinMax/Merge/IsNeg/IsPos/PMap/Range/Reverse/Minus/Some/IsSubset/IsSuperset/Take/TakeLast/Union/IsZero/Zip/GroupBy/UniqBy/Flatten/Prepend/Partition/Tail/Head/SplitEvery)
	*	Credit: https://github.com/logic-building/functional-go
	* Credit: https://github.com/achannarasappa/pneumatic
**/

type fnObj func(interface{}) interface{}

// Transformer Define Transformer Pattern interface
type Transformer[T any, R any] interface {
	TransformedBy() TransformerFunctor[T, R]
}

// TransformerFunctor Functor of Transform
type TransformerFunctor[T any, R any] func(T) R

// ReducerFunctor Functor for Reduce
type ReducerFunctor[T any, R any] func(R, T) R

// Predicate Predicate Functor
type Predicate[T any] func(T) bool

// PredicateErr Predicate Functor
type PredicateErr[T any] func(T, int) (bool, error)

// Comparator Comparator Functor
type Comparator[T any] func(T, T) bool

// Comparable Comparable interface able to be compared
type Comparable[T any] interface {
	CompareTo(T) int
}

// Numeric Define Numeric types for Generics
type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// Ordered Define Ordered types for Generics
type Ordered interface {
	int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | uintptr | string | float32 | float64
}

// CompareToOrdered A general Compare function for Ordered
func CompareToOrdered[T Ordered](a, b T) int {
	if b > a {
		return 1
	} else if b < a {
		return -1
	}

	return 0
}

// PMapOption Options for PMap usages
type PMapOption struct {
	FixedPool   int // number of goroutines
	RandomOrder bool
}

// Compose Compose the functions from right to left (Math: f(g(x)) Compose: Compose(f, g)(x))
func Compose[T any](fnList ...func(...T) []T) func(...T) []T {
	return func(s ...T) []T {
		f := fnList[0]
		nextFnList := fnList[1:]

		if len(fnList) == 1 {
			return f(s...)
		}

		return f(Compose(nextFnList...)(s...)...)
	}
}

// ComposeInterface Compose the functions from right to left (Math: f(g(x)) Compose: Compose(f, g)(x))
func ComposeInterface(fnList ...func(...interface{}) []interface{}) func(...interface{}) []interface{} {
	return Compose(fnList...)
}

// Pipe Pipe the functions from left to right
func Pipe[T any](fnList ...func(...T) []T) func(...T) []T {
	return func(s ...T) []T {

		lastIndex := len(fnList) - 1
		f := fnList[lastIndex]
		nextFnList := fnList[:lastIndex]

		if len(fnList) == 1 {
			return f(s...)
		}

		return f(Pipe(nextFnList...)(s...)...)
	}
}

// PipeInterface Pipe the functions from left to right
func PipeInterface(fnList ...func(...interface{}) []interface{}) func(...interface{}) []interface{} {
	return Pipe(fnList...)
}

// Map Map the values to the function from left to right
func Map[T any, R any](fn TransformerFunctor[T, R], values ...T) []R {
	result := make([]R, len(values))
	for i, val := range values {
		result[i] = fn(val)
	}

	return result
}

// MapIndexed Map the values to the function from left to right
func MapIndexed[T any, R any](fn func(T, int) R, values ...T) []R {
	result := make([]R, len(values))
	for i, val := range values {
		result[i] = fn(val, i)
	}

	return result
}

// Reduce Reduce the values from left to right(func(memo,val), starting value, slice)
func Reduce[T any, R any](fn ReducerFunctor[T, R], memo R, input ...T) R {

	for i := 0; i < len(input); i++ {
		memo = fn(memo, input[i])
	}

	return memo
}

// ReduceIndexed Reduce the values from left to right(func(memo,val,index), starting value, slice)
func ReduceIndexed[T any, R any](fn func(R, T, int) R, memo R, input ...T) R {

	for i := 0; i < len(input); i++ {
		memo = fn(memo, input[i], i)
	}

	return memo
}

// Filter Filter the values by the given predicate function (predicate func, slice)
func Filter[T any](fn func(T, int) bool, input ...T) []T {
	var list = make([]T, len(input))

	var newLen = 0

	for i := range input {
		if fn(input[i], i) {
			newLen++
			list[newLen-1] = input[i]
		}
	}

	result := list[:newLen]

	return result
}

// Reject Reject the values by the given predicate function (predicate func, slice)
func Reject[T any](fn func(T, int) bool, input ...T) []T {
	return Filter(func(val T, i int) bool {
		return !fn(val, i)
	}, input...)
}

// Concat Concat slices
func Concat[T any](mine []T, slices ...[]T) []T {
	var mineLen = len(mine)
	var totalLen = mineLen

	for _, slice := range slices {
		if slice == nil {
			continue
		}

		var targetLen = len(slice)
		totalLen += targetLen
	}
	var newOne = make([]T, totalLen)

	for i, item := range mine {
		newOne[i] = item
	}
	totalIndex := mineLen

	for _, slice := range slices {
		if slice == nil {
			continue
		}

		var target = slice
		var targetLen = len(target)
		for j, item := range target {
			newOne[totalIndex+j] = item
		}
		totalIndex += targetLen
	}

	return newOne
}

// SortSlice Sort items by Comparator
func SortSlice[T any](fn Comparator[T], input ...T) []T {
	Sort(fn, input)

	return input
}

// SortOrderedAscending Sort items by Comparator
func SortOrderedAscending[T Ordered](input ...T) []T {
	SortOrdered(true, input...)

	return input
}

// SortOrderedDescending Sort items by Comparator
func SortOrderedDescending[T Ordered](input ...T) []T {
	SortOrdered(false, input...)

	return input
}

// SortOrdered Sort items by Comparator
func SortOrdered[T Ordered](ascending bool, input ...T) []T {
	if ascending {
		Sort(func(a, b T) bool {
			return CompareToOrdered(a, b) > 0
		}, input)
	} else {
		Sort(func(a, b T) bool {
			return CompareToOrdered(a, b) < 0
		}, input)
	}

	return input
}

// Sort Sort items by Comparator
func Sort[T any](fn Comparator[T], input []T) {
	sort.SliceStable(input, func(previous int, next int) bool {
		return fn(input[previous], input[next])
	})
}

// Dedupe Returns a new list removing consecutive duplicates in list.
func Dedupe[T comparable](list ...T) []T {
	var newList []T

	lenList := len(list)
	for i := 0; i < lenList; i++ {
		if i+1 < lenList && list[i] == list[i+1] {
			continue
		}
		newList = append(newList, list[i])
	}
	return newList
}

// Difference returns a set that is the first set without elements of the remaining sets
// repeated value within list parameter will be ignored
func Difference[T comparable](arrList ...[]T) []T {
	if arrList == nil {
		return make([]T, 0)
	}

	resultMap := make(map[T]interface{})
	if len(arrList) == 1 {
		return Distinct(arrList[0]...)
	}

	var newList []T
	// 1st loop iterates items in 1st array
	// 2nd loop iterates all the rest of the arrays
	// 3rd loop iterates items in the rest of the arrays
	for i := 0; i < len(arrList[0]); i++ {

		matchCount := 0
		for j := 1; j < len(arrList); j++ {
			for _, v := range arrList[j] {
				// compare every items in 1st array to every items in the rest of the arrays
				if arrList[0][i] == v {
					matchCount++
					break
				}
			}
		}
		if matchCount == 0 {
			_, ok := resultMap[arrList[0][i]]
			if !ok {
				newList = append(newList, arrList[0][i])
				resultMap[arrList[0][i]] = true
			}
		}
	}
	return newList
}

// Distinct removes duplicates.
//
// Example
// 	list := []int{8, 2, 8, 0, 2, 0}
// 	Distinct(list...) // returns [8, 2, 0]
func Distinct[T comparable](list ...T) []T {
	// Keep order
	resultIndex := 0
	maxLen := len(list)
	result := make([]T, maxLen)
	if maxLen > 0 {
		s := make(map[T]bool)

		for _, v := range list {
			if !s[v] {
				result[resultIndex] = v
				s[v] = true

				resultIndex++
			}
		}

		return result[:resultIndex]
	}

	return result
}

// DistinctForInterface removes duplicates.
//
// Example
// 	list := []interface{}{8, 2, 8, 0, 2, 0}
// 	DistinctForInterface(list...) // returns [8, 2, 0]
func DistinctForInterface(list ...interface{}) []interface{} {
	// Keep order
	resultIndex := 0
	maxLen := len(list)
	result := make([]interface{}, maxLen)
	if maxLen > 0 {
		s := make(map[interface{}]bool)

		for _, v := range list {
			if !s[v] {
				result[resultIndex] = v
				s[v] = true

				resultIndex++
			}
		}

		return result[:resultIndex]
	}

	return result
}

// DistinctRandom removes duplicates.(RandomOrder)
func DistinctRandom[T comparable](list ...T) []T {
	s := SliceToMap(true, list...)
	return Keys(s)
}

// IsDistinct returns true if no two of the arguments are =
func IsDistinct[T comparable](list ...T) bool {
	if len(list) == 0 {
		return false
	}

	s := make(map[T]bool)
	for _, v := range list {
		if _, ok := s[v]; ok {
			return false
		}
		s[v] = true
	}
	return true
}

// DropEq returns a new list after dropping the given item
//
// Example:
// 	DropEq(1, 1, 2, 3, 1) // returns [2, 3]
func DropEq[T comparable](num T, list ...T) []T {
	var newList []T
	for _, v := range list {
		if v != num {
			newList = append(newList, v)
		}
	}
	return newList
}

// Drop drops N item(s) from the list and returns new list.
// Returns empty list if there is only one item in the list or list empty
func Drop[T any](count int, list ...T) []T {
	if count <= 0 {
		return list
	}

	if count >= len(list) {
		return make([]T, 0)
	}

	return list[count:]
}

// DropLast drops last N item(s) from the list and returns new list.
// Returns empty list if there is only one item in the list or list empty
func DropLast[T any](count int, list ...T) []T {
	listLen := len(list)

	if listLen == 0 || count >= listLen {
		return make([]T, 0)
	}

	return list[:(listLen - count)]
}

// DropWhile drops the items from the list as long as condition satisfies.
//
// Takes two inputs
//	1. Function: takes one input and returns boolean
//	2. list
//
// Returns:
// 	New List.
//  Empty list if either one of arguments or both of them are nil
//
// Example: Drops even number. Returns the remaining items once odd number is found in the list.
//	DropWhile(isEven, 4, 2, 3, 4, 5) // Returns [3, 4, 5]
//
//	func isEven(num int) bool {
//		return num%2 == 0
//	}
func DropWhile[T any](f Predicate[T], list ...T) []T {
	if f == nil {
		return make([]T, 0)
	}
	var newList []T
	for i, v := range list {
		if !f(v) {
			listLen := len(list)
			newList = make([]T, listLen-i)
			j := 0
			for i < listLen {
				newList[j] = list[i]
				i++
				j++
			}
			return newList
		}
	}
	return newList
}

// IsEqual Returns true if both list are equal else returns false
func IsEqual[T comparable](list1, list2 []T) bool {
	len1 := len(list1)
	len2 := len(list2)

	if len1 == 0 || len2 == 0 || len1 != len2 {
		return false
	}

	for i := 0; i < len1; i++ {
		if list1[i] != list2[i] {
			return false
		}
	}
	return true
}

// IsEqualMap Returns true if both maps are equal else returns false
func IsEqualMap[T comparable, R comparable](map1, map2 map[T]R) bool {
	len1 := len(map1)
	len2 := len(map2)

	if len1 == 0 || len2 == 0 || len1 != len2 {
		return false
	}

	for k1, v1 := range map1 {
		found := false
		for k2, v2 := range map2 {
			if k1 == k2 && v1 == v2 {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

// // IsEven Returns true if n is even
// func IsEven[T Numeric](v T) bool {
// 	return v%2 == 0
// }
//
// // IsOdd Returns true if n is odd
// func IsOdd[T Numeric](v T) bool {
// 	return v%2 != 0
// }

// Every returns true if supplied function returns logical true for every item in the list
//
// Example:
//	Every(even, 8, 2, 10, 4) // Returns true
//
//	func isEven(num int) bool {
//		return num%2 == 0
//	}
//
// Every(even) // Returns false
// Every(nil) // Returns false
func Every[T any](f Predicate[T], list ...T) bool {
	if f == nil || len(list) == 0 {
		return false
	}
	for _, v := range list {
		if !f(v) {
			return false
		}
	}
	return true
}

// Exists checks if given item exists in the list
//
// Example:
//	Exists(8, 8, 2, 10, 4) // Returns true
//	Exists(8) // Returns false
func Exists[T comparable](input T, list ...T) bool {
	for _, v := range list {
		if v == input {
			return true
		}
	}
	return false
}

// ExistsForInterface checks if given item exists in the list
//
// Example:
//	ExistsForInterface(8, 8, 2, 10, 4) // Returns true
//	ExistsForInterface(8) // Returns false
func ExistsForInterface(input interface{}, list ...interface{}) bool {
	for _, v := range list {
		if v == input {
			return true
		}
	}
	return false
}

// Intersection return a set that is the intersection of the input sets
// repeated value within list parameter will be ignored
func Intersection[T comparable](inputList ...[]T) []T {
	inputLen := len(inputList)
	if inputList == nil {
		return make([]T, 0)
	}

	if inputLen == 1 {
		resultMap := make(map[T]interface{}, len(inputList[0]))
		var newList []T
		for i := 0; i < len(inputList[0]); i++ {
			_, ok := resultMap[inputList[0][i]]
			if !ok {
				newList = append(newList, inputList[0][i])
				resultMap[inputList[0][i]] = true
			}
		}
		return newList
	}

	resultMap := make(map[T]interface{})
	var newList []T
	// 1st loop iterates items in 1st array
	// 2nd loop iterates all the rest of the arrays
	// 3rd loop iterates items in the rest of the arrays
	for i := 0; i < len(inputList[0]); i++ {

		matchCount := 0
		for j := 1; j < inputLen; j++ {
			for _, v := range inputList[j] {
				// compare every items in 1st array to every items in the rest of the arrays
				if inputList[0][i] == v {
					matchCount++
					break
				}
			}
		}
		if matchCount == inputLen-1 {
			_, ok := resultMap[inputList[0][i]]
			if !ok {
				newList = append(newList, inputList[0][i])
				resultMap[inputList[0][i]] = true
			}
		}
	}
	return newList
}

// IntersectionForInterface return a set that is the intersection of the input sets
// repeated value within list parameter will be ignored
func IntersectionForInterface(inputList ...[]interface{}) []interface{} {
	inputLen := len(inputList)
	if inputList == nil {
		return make([]interface{}, 0)
	}

	if inputLen == 1 {
		resultMap := make(map[interface{}]interface{}, len(inputList[0]))
		var newList []interface{}
		for i := 0; i < len(inputList[0]); i++ {
			_, ok := resultMap[inputList[0][i]]
			if !ok {
				newList = append(newList, inputList[0][i])
				resultMap[inputList[0][i]] = true
			}
		}
		return newList
	}

	resultMap := make(map[interface{}]interface{})
	var newList []interface{}
	// 1st loop iterates items in 1st array
	// 2nd loop iterates all the rest of the arrays
	// 3rd loop iterates items in the rest of the arrays
	for i := 0; i < len(inputList[0]); i++ {

		matchCount := 0
		for j := 1; j < inputLen; j++ {
			for _, v := range inputList[j] {
				// compare every items in 1st array to every items in the rest of the arrays
				if inputList[0][i] == v {
					matchCount++
					break
				}
			}
		}
		if matchCount == inputLen-1 {
			_, ok := resultMap[inputList[0][i]]
			if !ok {
				newList = append(newList, inputList[0][i])
				resultMap[inputList[0][i]] = true
			}
		}
	}
	return newList
}

// IntersectionMapByKey return a set that is the intersection of the input sets
func IntersectionMapByKey[T comparable, R any](inputList ...map[T]R) map[T]R {
	inputLen := len(inputList)

	if inputLen == 0 {
		return make(map[T]R)
	}

	if inputLen == 1 {
		resultMap := make(map[T]R, len(inputList[0]))
		for k, v := range inputList[0] {
			resultMap[k] = v
		}
		return resultMap
	}

	resultMap := make(map[T]R)
	countMap := make(map[T]int)
	for _, mapItem := range inputList {
		for k, v := range mapItem {
			_, exists := resultMap[k]
			if !exists {
				resultMap[k] = v
			}
			countMap[k]++
		}
	}
	for k, v := range countMap {
		if v < inputLen {
			delete(resultMap, k)
		}
	}
	return resultMap
}

// IntersectionMapByKeyForInterface return a set that is the intersection of the input sets
func IntersectionMapByKeyForInterface[R any](inputList ...map[interface{}]R) map[interface{}]R {
	inputLen := len(inputList)

	if inputLen == 0 {
		return make(map[interface{}]R)
	}

	if inputLen == 1 {
		resultMap := make(map[interface{}]R, len(inputList[0]))
		for k, v := range inputList[0] {
			resultMap[k] = v
		}
		return resultMap
	}

	resultMap := make(map[interface{}]R)
	countMap := make(map[interface{}]int)
	for _, mapItem := range inputList {
		for k, v := range mapItem {
			_, exists := resultMap[k]
			if !exists {
				resultMap[k] = v
			}
			countMap[k]++
		}
	}
	for k, v := range countMap {
		if v < inputLen {
			delete(resultMap, k)
		}
	}
	return resultMap
}

// Minus all of set1 but not in set2
func Minus[T comparable](set1, set2 []T) []T {
	resultIndex := 0
	maxLen := len(set1)
	result := make([]T, maxLen)
	set2Map := SliceToMap(true, set2...)

	for _, item := range set1 {
		_, exists := set2Map[item]
		if !exists {
			result[resultIndex] = item
			resultIndex++
		}
	}

	return result[:resultIndex]
}

// MinusForInterface all of set1 but not in set2
func MinusForInterface(set1, set2 []interface{}) []interface{} {
	resultIndex := 0
	maxLen := len(set1)
	result := make([]interface{}, maxLen)
	set2Map := SliceToMapForInterface(true, set2...)

	for _, item := range set1 {
		_, exists := set2Map[item]
		if !exists {
			result[resultIndex] = item
			resultIndex++
		}
	}

	return result[:resultIndex]
}

// MinusMapByKey all of set1 but not in set2
func MinusMapByKey[T comparable, R any](set1, set2 map[T]R) map[T]R {
	resultMap := make(map[T]R, len(set1))

	for k, v := range set1 {
		_, exists := set2[k]
		if !exists {
			resultMap[k] = v
		}
	}

	return resultMap
}

// Keys returns a slice of map's keys
func Keys[T comparable, R any](m map[T]R) []T {
	keys := make([]T, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// KeysForInterface returns a slice of map's keys
func KeysForInterface[R any](m map[interface{}]R) []interface{} {
	keys := make([]interface{}, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// Values returns a slice of map's values
func Values[T comparable, R any](m map[T]R) []R {
	keys := make([]R, len(m))
	i := 0
	for _, v := range m {
		keys[i] = v
		i++
	}
	return keys
}

// ValuesForInterface returns a slice of map's values
func ValuesForInterface[R any](m map[interface{}]R) []R {
	keys := make([]R, len(m))
	i := 0
	for _, v := range m {
		keys[i] = v
		i++
	}
	return keys
}

// Max returns max item from the list.
// Return 0 if the list is either empty or nil
func Max[T Numeric](list ...T) T {
	if list == nil || len(list) == 0 {
		return 0
	}
	result := list[0]
	for _, v := range list {
		if v > result {
			result = v
		}
	}
	return result
}

// Min returns min item from the list.
// Return 0 if the list is either empty or nil
func Min[T Numeric](list ...T) T {
	if list == nil || len(list) == 0 {
		return 0
	}
	result := list[0]
	for _, v := range list {
		if v < result {
			result = v
		}
	}
	return result
}

// MinMax returns min and max items from the list.
// Return 0,0 if the list is either empty or nil
func MinMax[T Numeric](list ...T) (T, T) {
	if list == nil || len(list) == 0 {
		return 0, 0
	}
	min := list[0]
	max := list[0]

	for _, v := range list {
		if v < min {
			min = v
		} else if v > max {
			max = v
		}
	}
	return min, max
}

// Merge takes two inputs: map[T]R and map[T]R and merge two maps and returns a new map[T]R.
func Merge[T comparable, R any](map1, map2 map[T]R) map[T]R {
	if map1 == nil && map2 == nil {
		return map[T]R{}
	}

	newMap := make(map[T]R, len(map1)+len(map2))

	if map1 == nil {
		for k, v := range map2 {
			newMap[k] = v
		}
		return newMap
	}

	if map2 == nil {
		for k, v := range map1 {
			newMap[k] = v
		}
		return newMap
	}

	for k, v := range map1 {
		newMap[k] = v
	}

	for k, v := range map2 {
		newMap[k] = v
	}

	return newMap
}

// MergeForInterface takes two inputs: map[T]R and map[T]R and merge two maps and returns a new map[T]R.
func MergeForInterface[R any](map1, map2 map[interface{}]R) map[interface{}]R {
	if map1 == nil && map2 == nil {
		return map[interface{}]R{}
	}

	newMap := make(map[interface{}]R, len(map1)+len(map2))

	if map1 == nil {
		for k, v := range map2 {
			newMap[k] = v
		}
		return newMap
	}

	if map2 == nil {
		for k, v := range map1 {
			newMap[k] = v
		}
		return newMap
	}

	for k, v := range map1 {
		newMap[k] = v
	}

	for k, v := range map2 {
		newMap[k] = v
	}

	return newMap
}

// IsNeg Returns true if num is less than zero, else false
func IsNeg[T Numeric](v T) bool {
	if v < 0 {
		return true
	}
	return false
}

// IsPos Returns true if num is great than zero, else false
func IsPos[T Numeric](v T) bool {
	if v > 0 {
		return true
	}
	return false
}

// PMap applies the function(1st argument) on each item in the list and returns a new list.
//  Order of new list is guaranteed. This feature can be disabled by passing: PMapOption{RandomOrder: true} to gain performance
//  Run in parallel. no_of_goroutines = no_of_items_in_list or 3rd argument can be passed to fix the number of goroutines.
//
// Takes 3 inputs. 3rd argument is option
//  1. Function - takes 1 input
//  2. optional argument - PMapOption{FixedPool: <some_number>}
//  3. List
func PMap[T any, R any](f TransformerFunctor[T, R], option *PMapOption, list ...T) []R {
	if f == nil {
		return make([]R, 0)
	}

	var worker = len(list)
	if option != nil {
		if option.FixedPool > 0 && option.FixedPool < worker {
			worker = option.FixedPool
		}

		if option.RandomOrder == true {
			return pMapNoOrder(f, list, worker)
		}
	}

	return pMapPreserveOrder(f, list, worker)
}

func pMapPreserveOrder[T any, R any](f TransformerFunctor[T, R], list []T, worker int) []R {
	chJobs := make(chan map[int]T, len(list))
	go func() {
		for i, v := range list {
			chJobs <- map[int]T{i: v}
		}
		close(chJobs)
	}()

	chResult := make(chan map[int]R, worker/3)

	var wg sync.WaitGroup

	for i := 0; i < worker; i++ {
		wg.Add(1)

		go func(chResult chan map[int]R, chJobs chan map[int]T) {
			defer wg.Done()

			for m := range chJobs {
				for k, v := range m {
					chResult <- map[int]R{k: f(v)}
				}
			}
		}(chResult, chJobs)
	}

	// This will wait for the workers to complete their job and then close the channel
	go func() {
		wg.Wait()
		close(chResult)
	}()

	newListMap := make(map[int]R, len(list))
	newList := make([]R, len(list))

	for m := range chResult {
		for k, v := range m {
			newListMap[k] = v
		}
	}

	for i := 0; i < len(list); i++ {
		newList[i] = newListMap[i]
	}

	return newList
}

func pMapNoOrder[T any, R any](f TransformerFunctor[T, R], list []T, worker int) []R {
	chJobs := make(chan T, len(list))
	go func() {
		for _, v := range list {
			chJobs <- v
		}
		close(chJobs)
	}()

	chResult := make(chan R, worker/3)

	var wg sync.WaitGroup

	for i := 0; i < worker; i++ {
		wg.Add(1)

		go func(chResult chan R, chJobs chan T) {
			defer wg.Done()

			for v := range chJobs {
				chResult <- f(v)
			}
		}(chResult, chJobs)
	}

	// This will wait for the workers to complete their job and then close the channel
	go func() {
		wg.Wait()
		close(chResult)
	}()

	newList := make([]R, len(list))
	i := 0

	for v := range chResult {
		newList[i] = v
		i++
	}

	return newList
}

// Range returns a list of range between lower and upper value
//
// Takes 3 inputs
//	1. lower limit
//	2. Upper limit
//	3. Hops (optional)
//
// Returns
//	List of range between lower and upper value
//	Empty list if 3rd argument is either 0 or negative number
//
// Example:
//	Range(-2, 2) // Returns: [-2, -1, 0, 1]
//	Range(0, 2) // Returns: [0, 1]
//	Range(3, 7, 2) // Returns: [3, 5]
func Range[T Numeric](lower, higher T, hops ...T) []T {
	hop := T(1)
	if len(hops) > 0 {
		if hops[0] <= 0 {
			return make([]T, 0)
		}
		hop = hops[0]
	}

	if lower >= higher {
		return make([]T, 0)
	}

	var l []T
	for _, v := 0, lower; v < higher; v += hop {
		l = append(l, v)
	}
	return l
}

// Reverse reverse the list
func Reverse[T any](list ...T) []T {
	newList := make([]T, len(list))
	for i := 0; i < len(list); i++ {
		newList[i] = list[len(list)-(i+1)]
	}
	return newList
}

// Some finds item in the list based on supplied function.
//
// Takes 2 input:
//	1. Function
//	2. List
//
// Returns:
//	bool.
//	True if condition satisfies, else false
//
// Example:
//	Some(isEven, 8, 2, 10, 4) // Returns true
//	Some(isEven, 1, 3, 5, 7) // Returns false
//	Some(nil) // Returns false
//
//	func isEven(num int) bool {
//		return num%2 == 0
//	}
func Some[T any](f Predicate[T], list ...T) bool {
	if f == nil {
		return false
	}
	for _, v := range list {
		if f(v) {
			return true
		}
	}
	return false
}

// IsSubset returns true or false by checking if set1 is a subset of set2
// repeated value within list parameter will be ignored
func IsSubset[T comparable](list1, list2 []T) bool {
	if list1 == nil || len(list1) == 0 || list2 == nil || len(list2) == 0 {
		return false
	}

	resultMap := make(map[T]interface{})
	for i := 0; i < len(list1); i++ {
		_, ok := resultMap[list1[i]]
		if !ok {
			found := false
			resultMap[list1[i]] = true
			for j := 0; j < len(list2); j++ {
				if list1[i] == list2[j] {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

// IsSubsetForInterface returns true or false by checking if set1 is a subset of set2
// repeated value within list parameter will be ignored
func IsSubsetForInterface(list1, list2 []interface{}) bool {
	if list1 == nil || len(list1) == 0 || list2 == nil || len(list2) == 0 {
		return false
	}

	resultMap := make(map[interface{}]interface{})
	for i := 0; i < len(list1); i++ {
		_, ok := resultMap[list1[i]]
		if !ok {
			found := false
			resultMap[list1[i]] = true
			for j := 0; j < len(list2); j++ {
				if list1[i] == list2[j] {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

// IsSuperset returns true or false by checking if set1 is a superset of set2
// repeated value within list parameter will be ignored
func IsSuperset[T comparable](list1, list2 []T) bool {
	return IsSubset(list2, list1)
}

// IsSupersetForInterface returns true or false by checking if set1 is a superset of set2
// repeated value within list parameter will be ignored
func IsSupersetForInterface(list1, list2 []interface{}) bool {
	return IsSubsetForInterface(list2, list1)
}

// IsSubsetMapByKey returns true or false by checking if set1 is a subset of set2
func IsSubsetMapByKey[T comparable, R any](item1, item2 map[T]R) bool {
	if item1 == nil || len(item1) == 0 || item2 == nil || len(item2) == 0 {
		return false
	}

	for k1 := range item1 {
		_, found := item2[k1]
		if !found {
			return false
		}
	}
	return true
}

// IsSubsetMapByKeyForInterface returns true or false by checking if set1 is a subset of set2
func IsSubsetMapByKeyForInterface[R any](item1, item2 map[interface{}]R) bool {
	if item1 == nil || len(item1) == 0 || item2 == nil || len(item2) == 0 {
		return false
	}

	for k1 := range item1 {
		_, found := item2[k1]
		if !found {
			return false
		}
	}
	return true
}

// IsSupersetMapByKey returns true or false by checking if set1 is a superset of set2
func IsSupersetMapByKey[T comparable, R any](item1, item2 map[T]R) bool {
	return IsSubsetMapByKey(item2, item1)
}

// IsSupersetMapByKeyForInterface returns true or false by checking if set1 is a superset of set2
func IsSupersetMapByKeyForInterface[R any](item1, item2 map[interface{}]R) bool {
	return IsSubsetMapByKeyForInterface(item2, item1)
}

// Take returns the first n elements of the slice
func Take[T any](count int, list ...T) []T {
	if count >= len(list) || count <= 0 {
		return list
	}

	return list[:count]
}

// TakeLast returns the last n elements of the slice
func TakeLast[T any](count int, list ...T) []T {
	listLen := len(list)

	if count >= listLen || count <= 0 {
		return list
	}

	return list[(listLen - count):]
}

// Union return a set that is the union of the input sets
// repeated value within list parameter will be ignored
func Union[T comparable](arrList ...[]T) []T {
	resultMap := make(map[T]interface{})
	for _, arr := range arrList {
		for _, v := range arr {
			resultMap[v] = true
		}
	}

	result := make([]T, len(resultMap))
	i := 0
	for k := range resultMap {
		result[i] = k
		i++
	}
	return result
}

// IsZero Returns true if num is zero, else false
func IsZero[T Numeric](v T) bool {
	if v == 0 {
		return true
	}
	return false
}

// Zip takes two inputs: first list of type: []T, second list of type: []T.
// Then it merges two list and returns a new map of type: map[T]R
func Zip[T comparable, R any](list1 []T, list2 []R) map[T]R {
	newMap := make(map[T]R)

	len1 := len(list1)
	len2 := len(list2)

	if len1 == 0 || len2 == 0 {
		return newMap
	}

	minLen := len1
	if len2 < minLen {
		minLen = len2
	}

	for i := 0; i < minLen; i++ {
		newMap[list1[i]] = list2[i]
	}

	return newMap
}

// GroupBy creates a map where the key is a group identifier and the value is a slice with the elements that have the same identifer
func GroupBy[T any, R comparable](grouper TransformerFunctor[T, R], list ...T) map[R][]T {
	var id R
	result := make(map[R][]T)
	for _, v := range list {
		id = grouper(v)
		result[id] = append(result[id], v)
	}
	return result
}

// UniqBy returns a slice of only unique values based on a comparable identifier
func UniqBy[T any, R comparable](identify TransformerFunctor[T, R], list ...T) []T {
	var id R
	result := make([]T, 0)
	identifiers := make(map[R]bool)
	for _, v := range list {
		id = identify(v)
		if _, ok := identifiers[id]; !ok {
			identifiers[id] = true
			result = append(result, v)
		}
	}
	return result
}

// Flatten creates a new slice where one level of nested elements are unnested
func Flatten[T any](list ...[]T) []T {

	result := make([]T, 0)

	// for _, v := range list {
	// 	result = append(result, v...)
	// }

	return Concat(result, list...)
}

// Prepend returns the slice with the additional element added to the beggining
func Prepend[T any](element T, list []T) []T {
	return append([]T{element}, list...)
}

// Partition splits elements into two groups - one where the predicate is satisfied and one where the predicate is not
func Partition[T any](predicate Predicate[T], list ...T) [][]T {
	resultTrue := make([]T, 0)
	resultFalse := make([]T, 0)
	for _, v := range list {
		if predicate(v) {
			resultTrue = append(resultTrue, v)
		} else {
			resultFalse = append(resultFalse, v)
		}
	}
	return [][]T{resultTrue, resultFalse}
}

// Tail returns the input slice with all elements except the first element
func Tail[T any](list ...T) []T {
	return Drop(1, list...)
}

// Head returns first element of a slice
func Head[T any](list ...T) T {
	var result T

	if len(list) <= 0 {
		return result
	}

	result = list[:1][0]

	return result
}

// SplitEvery returns elements in equal length slices
func SplitEvery[T any](size int, list ...T) [][]T {
	if size <= 0 || len(list) <= 1 {
		return [][]T{list}
	}

	result := make([][]T, 0)
	currentGroup := make([]T, 0)
	for i, v := range list {
		if len(currentGroup) < size {
			currentGroup = append(currentGroup, v)
		} else {
			result = append(result, currentGroup)
			currentGroup = []T{v}
		}

		if (i + 1) >= len(list) {
			result = append(result, currentGroup)
		}
	}

	return result
}

// Trampoline Trampoline
func Trampoline[T any](fn func(...T) ([]T, bool, error), input ...T) ([]T, error) {
	result := input
	var isDone bool
	var err error

	for {
		result, isDone, err = fn(result...)
		if err != nil {
			return nil, err
		}
		if isDone {
			break
		}
	}
	return result, err
}

// DuplicateSlice Return a new Slice
func DuplicateSlice[T any](list []T) []T {
	if len(list) > 0 {
		return append(list[:0:0], list...)
	}

	return make([]T, 0)
}

// DuplicateMap Return a new Map
func DuplicateMap[T comparable, R any](input map[T]R) map[T]R {
	if len(input) > 0 {
		newOne := make(map[T]R, len(input))
		for k, v := range input {
			newOne[k] = v
		}

		return newOne
	}

	return make(map[T]R)
}

// DuplicateMapForInterface Return a new Map
func DuplicateMapForInterface[R any](input map[interface{}]R) map[interface{}]R {
	if len(input) > 0 {
		newOne := make(map[interface{}]R, len(input))
		for k, v := range input {
			newOne[k] = v
		}

		return newOne
	}

	return make(map[interface{}]R)
}

// IsNil Check is it nil
func IsNil(obj interface{}) bool {
	val := reflect.ValueOf(obj)

	if Kind(obj) == reflect.Ptr {
		return val.IsNil()
	}
	return !val.IsValid()
}

// IsPtr Check is it a Ptr
func IsPtr(obj interface{}) bool {
	return Kind(obj) == reflect.Ptr
}

// Kind Get Kind by reflection
func Kind(obj interface{}) reflect.Kind {
	return reflect.ValueOf(obj).Kind()
}

// PtrOf Return Ptr of a value
func PtrOf[T any](v T) *T {
	return &v
}

// SliceOf Return Slice of varargs
func SliceOf[T any](args ...T) []T {
	return args
}

// SliceToMap Return Slice of varargs
func SliceToMap[T comparable, R any](defaultValue R, input ...T) map[T]R {
	resultMap := make(map[T]R)
	for _, key := range input {
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = defaultValue
		}
	}
	return resultMap
}

// SliceToMapForInterface Return Slice of varargs
func SliceToMapForInterface[R any](defaultValue R, input ...interface{}) map[interface{}]R {
	resultMap := make(map[interface{}]R)
	for _, key := range input {
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = defaultValue
		}
	}
	return resultMap
}

// MakeNumericReturnForVariadicParamReturnBool1 Make Numeric 1 bool Return (for compose() general fp functions simply)
func MakeNumericReturnForVariadicParamReturnBool1[T any, R Numeric](fn func(...T) bool) func(...T) []R {
	return func(args ...T) []R {
		if fn(args...) {
			return SliceOf(R(1))
		}

		return SliceOf(R(0))
	}
}

// MakeNumericReturnForSliceParamReturnBool1 Make Numeric 1 bool Return (for compose() general fp functions simply)
func MakeNumericReturnForSliceParamReturnBool1[T any, R Numeric](fn func([]T) bool) func(...T) []R {
	return func(args ...T) []R {
		if fn(args) {
			return SliceOf(R(1))
		}

		return SliceOf(R(0))
	}
}

// MakeNumericReturnForParam1ReturnBool1 Make Numeric 1 bool Return (for compose() general fp functions simply)
func MakeNumericReturnForParam1ReturnBool1[T any, R Numeric](fn func(T) bool) func(...T) []R {
	return func(args ...T) []R {
		if fn(args[0]) {
			return SliceOf(R(1))
		}

		return SliceOf(R(0))
	}
}

// MakeVariadicParam1 MakeVariadic for 1 Param (for compose() general fp functions simply)
func MakeVariadicParam1[T any, R any](fn func(T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0])
	}
}

// MakeVariadicParam2 MakeVariadic for 2 Params (for compose() general fp functions simply)
func MakeVariadicParam2[T any, R any](fn func(T, T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0], args[1])
	}
}

// MakeVariadicParam3 MakeVariadic for 3 Params (for compose() general fp functions simply)
func MakeVariadicParam3[T any, R any](fn func(T, T, T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0], args[1], args[2])
	}
}

// MakeVariadicParam4 MakeVariadic for 4 Params (for compose() general fp functions simply)
func MakeVariadicParam4[T any, R any](fn func(T, T, T, T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0], args[1], args[2], args[3])
	}
}

// MakeVariadicParam5 MakeVariadic for 5 Params (for compose() general fp functions simply)
func MakeVariadicParam5[T any, R any](fn func(T, T, T, T, T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0], args[1], args[2], args[3], args[4])
	}
}

// MakeVariadicParam6 MakeVariadic for 6 Params (for compose() general fp functions simply)
func MakeVariadicParam6[T any, R any](fn func(T, T, T, T, T, T) []R) func(...T) []R {
	return func(args ...T) []R {
		return fn(args[0], args[1], args[2], args[3], args[4], args[5])
	}
}

// MakeVariadicReturn1 MakeVariadic for 1 Return value (for compose() general fp functions simply)
func MakeVariadicReturn1[T any, R any](fn func(...T) R) func(...T) []R {
	return func(args ...T) []R {
		return []R{fn(args...)}
	}
}

// MakeVariadicReturn2 MakeVariadic for 2 Return values (for compose() general fp functions simply)
func MakeVariadicReturn2[T any, R any](fn func(...T) (R, R)) func(...T) []R {
	return func(args ...T) []R {
		r1, r2 := fn(args...)
		return []R{r1, r2}
	}
}

// MakeVariadicReturn3 MakeVariadic for 3 Return values (for compose() general fp functions simply)
func MakeVariadicReturn3[T any, R any](fn func(...T) (R, R, R)) func(...T) []R {
	return func(args ...T) []R {
		r1, r2, r3 := fn(args...)
		return []R{r1, r2, r3}
	}
}

// MakeVariadicReturn4 MakeVariadic for 4 Return values (for compose() general fp functions simply)
func MakeVariadicReturn4[T any, R any](fn func(...T) (R, R, R, R)) func(...T) []R {
	return func(args ...T) []R {
		r1, r2, r3, r4 := fn(args...)
		return []R{r1, r2, r3, r4}
	}
}

// MakeVariadicReturn5 MakeVariadic for 5 Return values (for compose() general fp functions simply)
func MakeVariadicReturn5[T any, R any](fn func(...T) (R, R, R, R, R)) func(...T) []R {
	return func(args ...T) []R {
		r1, r2, r3, r4, r5 := fn(args...)
		return []R{r1, r2, r3, r4, r5}
	}
}

// MakeVariadicReturn6 MakeVariadic for 6 Return values (for compose() general fp functions simply)
func MakeVariadicReturn6[T any, R any](fn func(...T) (R, R, R, R, R, R)) func(...T) []R {
	return func(args ...T) []R {
		r1, r2, r3, r4, r5, r6 := fn(args...)
		return []R{r1, r2, r3, r4, r5, r6}
	}
}

// CurryParam1ForSlice1 Curry for 1 Param (for currying general fp functions simply)
func CurryParam1ForSlice1[T any, R any, A any](fn func(A, []T) R, a A) func(...T) R {
	return func(args ...T) R {
		return fn(a, args)
	}
}

// CurryParam1 Curry for 1 Param (for currying general fp functions simply)
func CurryParam1[T any, R any, A any](fn func(A, ...T) R, a A) func(...T) R {
	return func(args ...T) R {
		return fn(a, args...)
	}
}

// CurryParam2 Curry for 2 Params (for currying general fp functions simply)
func CurryParam2[T any, R any, A any, B any](fn func(A, B, ...T) R, a A, b B) func(...T) R {
	return func(args ...T) R {
		return fn(a, b, args...)
	}
}

// CurryParam3 Curry for 3 Params (for currying general fp functions simply)
func CurryParam3[T any, R any, A any, B any, C any](fn func(A, B, C, ...T) R, a A, b B, c C) func(...T) R {
	return func(args ...T) R {
		return fn(a, b, c, args...)
	}
}

// CurryParam4 Curry for 4 Params (for currying general fp functions simply)
func CurryParam4[T any, R any, A any, B any, C any, D any](fn func(A, B, C, D, ...T) R, a A, b B, c C, d D) func(...T) R {
	return func(args ...T) R {
		return fn(a, b, c, d, args...)
	}
}

// CurryParam5 Curry for 5 Params (for currying general fp functions simply)
func CurryParam5[T any, R any, A any, B any, C any, D any, E any](fn func(A, B, C, D, E, ...T) R, a A, b B, c C, d D, e E) func(...T) R {
	return func(args ...T) R {
		return fn(a, b, c, d, e, args...)
	}
}

// CurryParam6 Curry for 6 Params (for currying general fp functions simply)
func CurryParam6[T any, R any, A any, B any, C any, D any, E any, F any](fn func(A, B, C, D, E, F, ...T) R, a A, b B, c C, d D, e E, f F) func(...T) R {
	return func(args ...T) R {
		return fn(a, b, c, d, e, f, args...)
	}
}

// CurryDef Curry inspired by Currying in Java ways
type CurryDef[T any, R any] struct {
	fn     func(c *CurryDef[T, R], args ...T) R
	result R
	isDone AtomBool

	callM sync.Mutex
	args  []T
}

// CurryNew New Curry instance by function
func CurryNew(fn func(c *CurryDef[interface{}, interface{}], args ...interface{}) interface{}) *CurryDef[interface{}, interface{}] {
	return CurryNewGenerics(fn)
}

// CurryNewGenerics New Curry instance by function
func CurryNewGenerics[T any, R any](fn func(c *CurryDef[T, R], args ...T) R) *CurryDef[T, R] {
	c := &CurryDef[T, R]{fn: fn}

	return c
}

// Call Call the currying function by partial or all args
func (currySelf *CurryDef[T, R]) Call(args ...T) *CurryDef[T, R] {
	currySelf.callM.Lock()
	if !currySelf.isDone.Get() {
		currySelf.args = append(currySelf.args, args...)
		currySelf.result = currySelf.fn(currySelf, currySelf.args...)
	}
	currySelf.callM.Unlock()
	return currySelf
}

// MarkDone Mark the currying is done(let others know it)
func (currySelf *CurryDef[T, R]) MarkDone() {
	currySelf.isDone.Set(true)
}

// IsDone Is the currying done
func (currySelf *CurryDef[T, R]) IsDone() bool {
	return currySelf.isDone.Get()
}

// Result Get the result value of currying
func (currySelf *CurryDef[T, R]) Result() R {
	return currySelf.result
}

// // Curry Curry utils instance
// var Curry CurryDef[interface{}, interface{}]

// PatternMatching

// Pattern Pattern general interface
type Pattern interface {
	Matches(value interface{}) bool
	Apply(interface{}) interface{}
}

// PatternMatching PatternMatching contains Pattern list
type PatternMatching struct {
	patterns []Pattern
}

// KindPatternDef Pattern which matching when the kind matches
type KindPatternDef struct {
	kind   reflect.Kind
	effect fnObj
}

// CompTypePatternDef Pattern which matching when the SumType matches
type CompTypePatternDef struct {
	compType CompType
	effect   fnObj
}

// EqualPatternDef Pattern which matching when the given object is equal to predefined one
type EqualPatternDef struct {
	value  interface{}
	effect fnObj
}

// RegexPatternDef Pattern which matching when the regex rule matches the given string
type RegexPatternDef struct {
	pattern string
	effect  fnObj
}

// OtherwisePatternDef Pattern which matching when the others didn't match(finally)
type OtherwisePatternDef struct {
	effect fnObj
}

// MatchFor Check does the given value match anyone of the Pattern list of PatternMatching
func (patternMatchingSelf PatternMatching) MatchFor(inValue interface{}) interface{} {
	for _, pattern := range patternMatchingSelf.patterns {
		value := inValue
		maybe := Maybe.Just(inValue)
		if maybe.IsKind(reflect.Ptr) {
			ptr := maybe.ToPtr()
			if reflect.TypeOf(*ptr).Kind() == (reflect.TypeOf(CompData{}).Kind()) {
				value = *ptr
			}
		}

		if pattern.Matches(value) {
			return pattern.Apply(value)
		}
	}

	panic(fmt.Sprintf("Cannot match %v", inValue))
}

// Matches Match the given value by the pattern
func (patternSelf KindPatternDef) Matches(value interface{}) bool {
	if Maybe.Just(value).IsNil() {
		return false
	}

	return patternSelf.kind == reflect.TypeOf(value).Kind()
}

// Matches Match the given value by the pattern
func (patternSelf CompTypePatternDef) Matches(value interface{}) bool {
	if Maybe.Just(value).IsPresent() && reflect.TypeOf(value).Kind() == reflect.TypeOf(CompData{}).Kind() {
		return MatchCompType(patternSelf.compType, (value).(CompData))
	}

	return patternSelf.compType.Matches(value)
}

// Matches Match the given value by the pattern
func (patternSelf EqualPatternDef) Matches(value interface{}) bool {
	return patternSelf.value == value
}

// Matches Match the given value by the pattern
func (patternSelf RegexPatternDef) Matches(value interface{}) bool {
	if Maybe.Just(value).IsNil() || reflect.TypeOf(value).Kind() != reflect.String {
		return false
	}

	matches, err := regexp.MatchString(patternSelf.pattern, (value).(string))
	if err == nil && matches {
		return true
	}

	return false
}

// Matches Match the given value by the pattern
func (patternSelf OtherwisePatternDef) Matches(value interface{}) bool {
	return true
}

// Apply Evaluate the result by its given effect function
func (patternSelf KindPatternDef) Apply(value interface{}) interface{} {
	return patternSelf.effect(value)
}

// Apply Evaluate the result by its given effect function
func (patternSelf CompTypePatternDef) Apply(value interface{}) interface{} {
	return patternSelf.effect(value)
}

// Apply Evaluate the result by its given effect function
func (patternSelf EqualPatternDef) Apply(value interface{}) interface{} {
	return patternSelf.effect(value)
}

// Apply Evaluate the result by its given effect function
func (patternSelf RegexPatternDef) Apply(value interface{}) interface{} {
	return patternSelf.effect(value)
}

// Apply Evaluate the result by its given effect function
func (patternSelf OtherwisePatternDef) Apply(value interface{}) interface{} {
	return patternSelf.effect(value)
}

// DefPattern Define the PatternMatching by Pattern list
func DefPattern(patterns ...Pattern) PatternMatching {
	return PatternMatching{patterns: patterns}
}

// InCaseOfKind In case of its Kind matches the given one
func InCaseOfKind(kind reflect.Kind, effect fnObj) Pattern {
	return KindPatternDef{kind: kind, effect: effect}
}

// InCaseOfSumType In case of its SumType matches the given one
func InCaseOfSumType(compType CompType, effect fnObj) Pattern {
	return CompTypePatternDef{compType: compType, effect: effect}
}

// InCaseOfEqual In case of its value is equal to the given one
func InCaseOfEqual(value interface{}, effect fnObj) Pattern {
	return EqualPatternDef{value: value, effect: effect}
}

// InCaseOfRegex In case of the given regex rule matches its value
func InCaseOfRegex(pattern string, effect fnObj) Pattern {
	return RegexPatternDef{pattern: pattern, effect: effect}
}

// Otherwise In case of the other patterns didn't match it
func Otherwise(effect fnObj) Pattern {
	return OtherwisePatternDef{effect: effect}
}

// Either Match Pattern list and return the effect() result of the matching Pattern
func Either(value interface{}, patterns ...Pattern) interface{} {
	return DefPattern(patterns...).MatchFor(value)
}

// SumType

// CompData Composite Data with values & its CompType(SumType)
type CompData struct {
	compType CompType
	objects  []interface{}
}

// CompType Abstract SumType concept interface
type CompType interface {
	Matches(value ...interface{}) bool
}

// SumType SumType contains a CompType list
type SumType struct {
	compTypes []CompType
}

// ProductType ProductType with a Kind list
type ProductType struct {
	kinds []reflect.Kind
}

// NilTypeDef NilType implemented by Nil determinations
type NilTypeDef struct {
}

// Matches Check does it match the SumType
func (typeSelf SumType) Matches(value ...interface{}) bool {
	for _, compType := range typeSelf.compTypes {
		if compType.Matches(value...) {
			return true
		}
	}

	return false
}

// Matches Check does it match the ProductType
func (typeSelf ProductType) Matches(value ...interface{}) bool {
	if len(value) != len(typeSelf.kinds) {
		return false
	}

	matches := true
	for i, v := range value {
		matches = matches && typeSelf.kinds[i] == Maybe.Just(v).Kind()
	}
	return matches
}

// Matches Check does it match nil
func (typeSelf NilTypeDef) Matches(value ...interface{}) bool {
	if len(value) != 1 {
		return false
	}

	return Maybe.Just(value[0]).IsNil()
}

// DefSum Define the SumType by CompType list
func DefSum(compTypes ...CompType) CompType {
	return SumType{compTypes: compTypes}
}

// DefProduct Define the ProductType of a SumType
func DefProduct(kinds ...reflect.Kind) CompType {
	return ProductType{kinds: kinds}
}

// NewCompData New SumType Data by its type and composite values
func NewCompData(compType CompType, value ...interface{}) *CompData {
	if compType.Matches(value...) {
		return &CompData{compType: compType, objects: value}
	}

	return nil
}

// MatchCompType Check does the Composite Data match the given SumType
func MatchCompType(compType CompType, value CompData) bool {
	return MatchCompTypeRef(compType, &value)
}

// MatchCompTypeRef Check does the Composite Data match the given SumType
func MatchCompTypeRef(compType CompType, value *CompData) bool {
	return compType.Matches(value.objects...)
}

// NilType NilType CompType instance
var NilType NilTypeDef

package fpgo

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
)

type fnObj func(interface{}) interface{}

// Transformer Define Transformer Pattern interface
type Transformer interface {
	TransformedBy() TransformerFunctor
}

// TransformerFunctor Functor of Transform
type TransformerFunctor func(interface{}) interface{}

// Predicate Predicate Functor
type Predicate func(interface{}) bool

// Comparator Comparator Functor
type Comparator func(interface{}, interface{}) bool

// Comparable Comparable interface able to be compared
type Comparable interface {
	CompareTo(interface{}) int
}

// Compose Compose the functions from right to left (Math: f(g(x)) Compose: Compose(f, g)(x))
func Compose(fnList ...func(...interface{}) []interface{}) func(...interface{}) []interface{} {
	return func(s ...interface{}) []interface{} {
		f := fnList[0]
		nextFnList := fnList[1:]

		if len(fnList) == 1 {
			return f(s...)
		}

		return f(Compose(nextFnList...)(s...)...)
	}
}

// Pipe Pipe the functions from left to right
func Pipe(fnList ...func(...interface{}) []interface{}) func(...interface{}) []interface{} {
	return func(s ...interface{}) []interface{} {

		lastIndex := len(fnList) - 1
		f := fnList[lastIndex]
		nextFnList := fnList[:lastIndex]

		if len(fnList) == 1 {
			return f(s...)
		}

		return f(Compose(nextFnList...)(s...)...)
	}
}

// Map Map the values to the function from left to right
func Map(fn TransformerFunctor, values ...interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, val := range values {
		result[i] = fn(val)
	}

	return result
}

// MapIndexed Map the values to the function from left to right
func MapIndexed(fn func(interface{}, int) interface{}, values ...interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, val := range values {
		result[i] = fn(val, i)
	}

	return result
}

// Reduce Reduce the values from left to right(func(memo,val), starting value, slice)
func Reduce(fn func(interface{}, interface{}) interface{}, memo interface{}, input ...interface{}) interface{} {

	for i := 0; i < len(input); i++ {
		memo = fn(memo, input[i])
	}

	return memo
}

// ReduceIndexed Reduce the values from left to right(func(memo,val,index), starting value, slice)
func ReduceIndexed(fn func(interface{}, interface{}, int) interface{}, memo interface{}, input ...interface{}) interface{} {

	for i := 0; i < len(input); i++ {
		memo = fn(memo, input[i], i)
	}

	return memo
}

// Filter Filter the values by the given predicate function (predicate func, slice)
func Filter(fn func(interface{}, int) bool, input ...interface{}) []interface{} {
	var list = make([]interface{}, len(input))

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
func Reject(fn func(interface{}, int) bool, input ...interface{}) []interface{} {
	return Filter(func(val interface{}, i int) bool {
		return !fn(val, i)
	}, input...)
}

// Concat Concat slices
func Concat(mine []interface{}, slices ...[]interface{}) []interface{} {
	var mineLen = len(mine)
	var totalLen = mineLen

	for _, slice := range slices {
		if slice == nil {
			continue
		}

		var targetLen = len(slice)
		totalLen += targetLen
	}
	var newOne = make([]interface{}, totalLen)

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
func SortSlice(fn Comparator, input ...interface{}) []interface{} {
	Sort(fn, input)

	return input
}

// Sort Sort items by Comparator
func Sort(fn Comparator, input []interface{}) {
	sort.SliceStable(input, func(previous int, next int) bool {
		return fn(input[previous], input[next])
	})
}

// PtrOf Return Ptr of a value
func PtrOf(v interface{}) *interface{} {
	return &v
}

// SliceOf Return Slice of varargs
func SliceOf(args ...interface{}) []interface{} {
	return args
}

// CurryDef Curry inspired by Currying in Java ways
type CurryDef struct {
	fn     func(c *CurryDef, args ...interface{}) interface{}
	result interface{}
	isDone AtomBool

	callM sync.Mutex
	args  []interface{}
}

// New New Curry instance by function
func (currySelf *CurryDef) New(fn func(c *CurryDef, args ...interface{}) interface{}) *CurryDef {
	c := &CurryDef{fn: fn}

	return c
}

// Call Call the currying function by partial or all args
func (currySelf *CurryDef) Call(args ...interface{}) *CurryDef {
	currySelf.callM.Lock()
	if !currySelf.isDone.Get() {
		currySelf.args = append(currySelf.args, args...)
		currySelf.result = currySelf.fn(currySelf, currySelf.args...)
	}
	currySelf.callM.Unlock()
	return currySelf
}

// MarkDone Mark the currying is done(let others know it)
func (currySelf *CurryDef) MarkDone() {
	currySelf.isDone.Set(true)
}

// IsDone Is the currying done
func (currySelf *CurryDef) IsDone() bool {
	return currySelf.isDone.Get()
}

// Result Get the result value of currying
func (currySelf *CurryDef) Result() interface{} {
	return currySelf.result
}

// Curry Curry utils instance
var Curry CurryDef

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

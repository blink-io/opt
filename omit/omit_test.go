package omit

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestConstruction(t *testing.T) {
	t.Parallel()

	hello := "hello"

	val := From("hello")
	checkState(t, val, StateSet)
	if !val.IsSet() {
		t.Error("should be set")
	}

	val = FromPtr(&hello)
	checkState(t, val, StateSet)
	val = FromPtr[string](nil)
	checkState(t, val, StateUnset)
	if !val.IsUnset() {
		t.Error("should be unset")
	}

	val = FromCond("hello", true)
	checkState(t, val, StateSet)
	val = FromCond("hello", false)
	checkState(t, val, StateUnset)
	if !val.IsUnset() {
		t.Error("should be unset")
	}

	val = Val[string]{}
	checkState(t, val, StateUnset)
	if !val.IsUnset() {
		t.Error("should be unset")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	val := From("hello")
	if val.MustGet() != "hello" {
		t.Error("wrong value")
	}
	if val.GetOr("hi") != "hello" {
		t.Error("wrong value")
	}
	if val.GetOrZero() != "hello" {
		t.Error("wrong value")
	}

	val.Unset()
	if _, ok := val.Get(); ok {
		t.Error("should not be okay")
	}
	if val.GetOr("hi") != "hi" {
		t.Error("wrong value")
	}
	if val.GetOrZero() != "" {
		t.Error("wrong value")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("should have panic'd")
		}
	}()
	_ = val.MustGet()
}

func TestMap(t *testing.T) {
	t.Parallel()

	val := Val[int]{}
	if !val.Map(func(int) int { return 0 }).IsUnset() {
		t.Error("it should still be unset")
	}
	if !Map(val, func(int) int { return 0 }).IsUnset() {
		t.Error("it should still be unset")
	}
	val.Set(5)
	if val.Map(func(i int) int { return i + 1 }).MustGet() != 6 {
		t.Error("wrong value")
	}
	if Map(val, func(i int) int { return i + 1 }).MustGet() != 6 {
		t.Error("wrong value")
	}
}

func TestChanges(t *testing.T) {
	t.Parallel()

	val := From("hello")
	checkState(t, val, StateSet)
	val.Unset()
	checkState(t, val, StateUnset)

	val = Val[string]{}
	checkState(t, val, StateUnset)
	val.Set("hello")
	checkState(t, val, StateSet)

	val = Val[string]{}
	checkState(t, val, StateUnset)
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	val := From("hello")
	checkJSON(t, val, `"hello"`)
	val.Unset()
	checkJSON(t, val, `null`)

}

func TestMarshalJSONIsZero(t *testing.T) {
	type testStruct struct {
		ID int
	}

	valSlice := Val[[]int]{}
	valSlice.Set(nil)
	if !valSlice.MarshalJSONIsZero() {
		t.Error("should be zero")
	}

	valMap := Val[map[string]int]{}
	valMap.Set(nil)
	if !valMap.MarshalJSONIsZero() {
		t.Error("should be zero")
	}

	valStruct := Val[*testStruct]{}
	valStruct.Set(nil)
	if !valStruct.MarshalJSONIsZero() {
		t.Error("should be zero")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	hello := Val[string]{}
	checkState(t, hello, StateUnset)

	if err := json.Unmarshal([]byte("null"), &hello); err == nil {
		t.Error("cannot accept a null")
	}

	if err := json.Unmarshal([]byte(`"hello"`), &hello); err != nil {
		t.Error(err)
	}
	checkState(t, hello, StateSet)

	if hello.MustGet() != "hello" {
		t.Error("expected hello")
	}

	hello.UnmarshalJSON(nil)
	checkState(t, hello, StateUnset)
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	hello := From("hello")
	b, err := hello.MarshalText()
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello" {
		t.Error("expected hello")
	}

	hello.Unset()
	b, err = hello.MarshalText()
	if err != nil {
		t.Error(err)
	}
	if string(b) != "" {
		t.Error("expected empty str")
	}

	marshaller := From(net.IPv4(1, 1, 1, 1))
	if b, err := marshaller.MarshalText(); err != nil {
		t.Error(err)
	} else if !bytes.Equal(b, []byte("1.1.1.1")) {
		t.Error("wrong value")
	}
}

func TestUnmarshalText(t *testing.T) {
	t.Parallel()

	var val Val[string]
	if err := val.UnmarshalText([]byte("")); err != nil {
		t.Error(err)
	}
	checkState(t, val, StateUnset)

	if err := val.UnmarshalText([]byte("hello")); err != nil {
		t.Error(err)
	}
	checkState(t, val, StateSet)
	if val.MustGet() != "hello" {
		t.Error("wrong value")
	}

	var unmarshaller Val[net.IP]
	if err := unmarshaller.UnmarshalText([]byte("")); err != nil {
		t.Error(err)
	}
	checkState(t, unmarshaller, StateUnset)

	if err := unmarshaller.UnmarshalText([]byte("1.1.1.1")); err != nil {
		t.Error(err)
	}
	checkState(t, unmarshaller, StateSet)
	if !unmarshaller.MustGet().Equal(net.IPv4(1, 1, 1, 1)) {
		t.Error("wrong value")
	}
}

func TestScan(t *testing.T) {
	t.Parallel()

	var val Val[string]
	if err := val.Scan(nil); err == nil {
		t.Error("should break trying to scan null")
	}

	if err := val.Scan("hello"); err != nil {
		t.Error(err)
	}
	checkState(t, val, StateSet)
	if val.MustGet() != "hello" {
		t.Error("wrong value")
	}
}

type valuerImplementation struct{}

func (valuerImplementation) Value() (driver.Value, error) {
	return int64(1), nil
}

func TestValue(t *testing.T) {
	t.Parallel()

	var val Val[string]
	if v, err := val.Value(); err != nil {
		t.Error(err)
	} else if v != nil {
		t.Error("expected v to be nil")
	}

	val = From("hello")
	if v, err := val.Value(); err != nil {
		t.Error(err)
	} else if v.(string) != "hello" {
		t.Error("expected v to be nil")
	}

	date := time.Date(2000, 1, 1, 2, 30, 0, 0, time.UTC)
	nullTime := From(date)
	if v, err := nullTime.Value(); err != nil {
		t.Error(err)
	} else if !v.(time.Time).Equal(date) {
		t.Error("time was wrong")
	}

	valuer := From(valuerImplementation{})
	if v, err := valuer.Value(); err != nil {
		t.Error(err)
	} else if v.(int64) != 1 {
		t.Error("expect const int")
	}
}

func TestStateStringer(t *testing.T) {
	t.Parallel()

	if StateUnset.String() != "unset" {
		t.Error("bad value")
	}
	if StateSet.String() != "set" {
		t.Error("bad value")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic")
		}
	}()
	_ = state(99).String()
}

func checkState[T any](t *testing.T, val Val[T], want state) {
	t.Helper()

	if want != val.State() {
		t.Errorf("state should be: %s but is: %s", want, val.State())
	}
}

func checkJSON[T any](t *testing.T, v Val[T], s string) {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Error(err)
	}

	if string(b) != s {
		t.Errorf("expect: %s, got: %s", s, b)
	}
}

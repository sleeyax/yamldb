package yamldb

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func newYamlDb() *YamlDb {
	return New(&DiskOptions{
		BasePath:        "test-data",
		AppendExtension: true,
	})
}

// Mock is a dummy struct that's only used for testing.
type Mock struct {
	Amount    int
	Ok        bool
	FooBarBaz string   `yaml:"foo"`
	Values    []string `yaml:",flow"`
}

func TestYamlDb_WriteRaw(t *testing.T) {
	db := New(&DiskOptions{
		BasePath:        "./test-data",
		AppendExtension: false,
	})
	defer db.PurgeAll()

	expected := []byte("bar")

	if err := db.WriteRaw("foo", expected); err != nil {
		t.Fatal(err)
	}

	actual, err := db.ReadRaw("foo")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(expected, actual) != 0 {
		t.Fatalf("Expected %v but got %v", expected, actual)
	}
}

func TestYamlDb_Write(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	m := Mock{
		Amount:    123,
		Ok:        true,
		FooBarBaz: "bar",
		Values:    []string{"a", "b", "c"},
	}

	if err := db.Write("test", m); err != nil {
		t.Fatal(err)
	}

	var out Mock
	if err := db.Read("test", &out); err != nil {
		t.Fatal(err)
	}

	if len(out.Values) != len(m.Values) {
		t.FailNow()
	}
}

func TestYamlDb_Delete(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	key := "foo"

	if err := db.WriteRaw(key, []byte("bar")); err != nil {
		t.Fatal(err)
	}

	if err := db.Delete(key); err != nil {
		t.Fatal(err)
	}

	if db.Has(key) {
		t.FailNow()
	}
}

func TestYamlDb_Update(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	err := db.Write("foo", Mock{
		Amount: 0,
		Ok:     false,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update("foo", &Mock{}, func(i interface{}) {
		u := i.(*Mock)
		u.Amount = 123
		u.Ok = true
	})
	if err != nil {
		t.Fatal(err)
	}

	var m Mock
	if err = db.Read("foo", &m); err != nil {
		t.Fatal(err)
	}

	if !m.Ok || m.Amount != 123 {
		t.FailNow()
	}
}

func TestYamlDb_Purge(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("data/value%d", i)
		if err := db.WriteRaw(key, []byte(key)); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Purge("data"); err != nil {
		t.Fatal(err)
	}

	if db.Has("data/value1") {
		t.FailNow()
	}
}

func TestYamlDb_Iterate(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	if err := db.WriteRaw("data/foo", []byte("bar")); err != nil {
		t.Fatal(err)
	}

	var count int

	ok, err := db.Iterate("data", func(key string, data []byte) error {
		count++
		return nil
	})

	if err != nil || !ok || count != 1 {
		t.FailNow()
	}
}

func TestYamlDb_IterateSerialized(t *testing.T) {
	db := newYamlDb()
	defer db.PurgeAll()

	if err := db.Write("data/foo", Mock{Ok: true}); err != nil {
		t.Fatal(err)
	}

	ok, err := db.IterateSerialized("data", &Mock{}, func(s interface{}) error {
		m := s.(*Mock)

		if !m.Ok {
			return errors.New("found mock is Ok = false")
		}

		return nil
	})

	if err != nil || !ok {
		t.FailNow()
	}
}

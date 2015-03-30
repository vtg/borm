package borm

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/vtg/pubsub"
)

var dbFile string
var db DB

func init() {
	_, filename, _, _ := runtime.Caller(0) // get full path of this file
	dbFile = path.Join(path.Dir(filename), "test.db")
	os.Remove(dbFile)
}

func openDB() {
	if !db.open {
		db, _ = Open(dbFile)
		// db.Log = true
	}
}

func assertEqual(t *testing.T, expect interface{}, v interface{}) {
	if !reflect.DeepEqual(v, expect) {
		_, fname, lineno, ok := runtime.Caller(1)
		if !ok {
			fname, lineno = "<UNKNOWN>", -1
		}
		t.Errorf("FAIL: %s:%d\nExpected: %#v\nReceived: %#v", fname, lineno, expect, v)
	}
}

func TestDBOpen(t *testing.T) {
	db, err := Open(dbFile)

	assertEqual(t, nil, err)
	assertEqual(t, dbFile, db.File)
	assertEqual(t, true, db.open)
}

func TestDBOpenError(t *testing.T) {
	db, err := Open("Z:::/qwe")

	assertEqual(t, "open Z:::/qwe: The filename, directory name, or volume label syntax is incorrect.", err.Error())
	assertEqual(t, "", db.File)
	assertEqual(t, false, db.open)
}

type Person struct {
	Model

	Name   string
	Active bool

	CreateTime
	UpdateTime
}

func Proc(t *testing.T, name string) func(e *pubsub.Event) {
	return func(e *pubsub.Event) {
		time.Sleep(1 * time.Second)
		assertEqual(t, name, e.Name)
		fmt.Println(e.Name, e.Objects)
	}
}

func TestEventName(t *testing.T) {
	p := Person{}
	assertEqual(t, "PersonCreated", eventName("Created", &p))
	assertEqual(t, "PersonDeleted", eventName("Deleted", &p))
}

func TestSave(t *testing.T) {
	openDB()

	// Events.Sub("PersonCreated", Proc(t, "PersonCreated"))
	// Events.Sub("PersonUpdated", Proc(t, "PersonUpdated"))

	// test creation
	p := Person{Name: "John Doe"}
	db.Save([]string{"people"}, &p)

	assertEqual(t, "1", p.ID)
	assertEqual(t, false, p.Active)

	// test update
	p.Active = true
	db.Save([]string{"people"}, &p)
	assertEqual(t, "1", p.ID)
	assertEqual(t, true, p.Active)
}

func TestFind(t *testing.T) {
	openDB()

	p := Person{Name: "John Doe"}
	db.Save([]string{"people"}, &p)

	assertEqual(t, true, p.ID != "")

	p1 := Person{}
	db.Find([]string{"people"}, p.ID, &p1)
	assertEqual(t, p.ID, p1.ID)
}

func TestList(t *testing.T) {
	openDB()

	p := Person{Name: "John Doe"}
	db.Save([]string{"peoplelist"}, &p)
	p1 := Person{Name: "John1 Doe"}
	db.Save([]string{"peoplelist"}, &p1)

	res := []Person{}
	db.List([]string{"peoplelist"}, &res)
	assertEqual(t, []Person{p, p1}, res)

	res1 := []*Person{}
	db.List([]string{"peoplelist"}, &res1)
	assertEqual(t, []*Person{&p, &p1}, res1)
}

func TestListKeys(t *testing.T) {
	openDB()

	p := Person{Name: "John Doe"}
	db.Save([]string{"list2"}, &p)
	p1 := Person{Name: "John1 Doe"}
	db.Save([]string{"list2"}, &p1)
	p2 := Person{Name: "John2 Doe"}
	db.Save([]string{"list2"}, &p2)
	p3 := Person{Name: "John3 Doe"}
	db.Save([]string{"list2"}, &p3)

	res := []Person{}
	db.ListKeys([]string{"list2"}, [][]byte{[]byte(p.ID), []byte(p3.ID), []byte("11111")}, &res)
	assertEqual(t, []Person{p, p3}, res)
}

func TestListItems(t *testing.T) {
	openDB()
	p := Person{Name: "John Doe"}
	db.Save([]string{"peoplelist1"}, &p)
	p1 := Person{Name: "John1 Doe"}
	db.Save([]string{"peoplelist1"}, &p1)

	res, _ := db.ListItems([]string{"peoplelist1"})
	assertEqual(t, 2, len(res))
}

func TestValues(t *testing.T) {
	openDB()
	p := Person{Name: "John Doe"}
	db.Save([]string{"list11"}, &p)
	p1 := Person{Name: "John1 Doe"}
	db.Save([]string{"list11"}, &p1)

	res, _ := db.Values([]string{"list11"})
	assertEqual(t, 2, len(res))
}

func TestDelete(t *testing.T) {
	openDB()

	// Events.Sub("PersonDeleted", Proc(t, "PersonDeleted"))

	p := Person{Name: "John Doe"}
	db.Save([]string{"people1"}, &p)

	db.Delete([]string{"people1"}, &p)

	p1 := Person{}

	db.Find([]string{"people1"}, p.ID, &p1)
	assertEqual(t, "", p1.ID)

	// time.Sleep(1 * time.Second)
}

func TestDeleteBucket(t *testing.T) {
	openDB()

	db.SaveValue([]string{"buck1", "buck2"}, "1", []byte("2"))

	err := db.DeleteBuckets([]string{"buck1", "buck2"}, []string{"1"})
	assertEqual(t, true, err != nil)

	err = db.DeleteBuckets([]string{"buck1"}, []string{"buck2"})
	assertEqual(t, nil, err)
}

func TestConcurrentCreation(t *testing.T) {
	openDB()
	var wg sync.WaitGroup
	db.Log = false

	for k := 0; k < 100; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := Person{Name: "John Doe"}
			db.Save([]string{"pep31"}, &p)
		}()
	}
	wg.Wait()

	res := []Person{}
	db.List([]string{"pep31"}, &res)
	assertEqual(t, 100, len(res))
}

func TestCount(t *testing.T) {
	openDB()
	var wg sync.WaitGroup

	db.Log = false
	for k := 0; k < 100; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := Person{Name: "John Doe"}
			db.Save([]string{"pep311"}, &p)
		}()
	}
	wg.Wait()

	res := db.Count([]string{"pep311"})
	assertEqual(t, 100, res)
}

var listFull bool

func benchListPrepare() {
	if listFull {
		return
	}
	openDB()
	db.Log = false
	for k := 0; k < 100; k++ {
		p := Person{Name: "John Doe"}
		db.Save([]string{"pep"}, &p)
	}

	listFull = true
}

func BenchmarkListItems(b *testing.B) {
	benchListPrepare()
	for n := 0; n < b.N; n++ {
		db.ListItems([]string{"pep"}, Params{Limit: 10})
	}
}

func BenchmarkListModels(b *testing.B) {
	benchListPrepare()
	r := []Person{}
	for n := 0; n < b.N; n++ {
		db.List([]string{"pep"}, &r, Params{Limit: 10})
	}
}

func BenchmarkListValues(b *testing.B) {
	benchListPrepare()
	for n := 0; n < b.N; n++ {
		db.Values([]string{"pep"}, Params{Limit: 10})
	}
}

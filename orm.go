package borm

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/vtg/pubsub"
)

// Events listener
var Events = *pubsub.Hub

// Params for List query
type Params struct {
	Offset  int
	Limit   int
	Reverse bool
}

func init() {
	Events.Start(4)
}

// DB
type DB struct {
	File string
	Log  bool

	db   *bolt.DB
	open bool
}

// Open opens database
func Open(dbfile string) (db DB, err error) {
	db.db, err = bolt.Open(dbfile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return
	}
	db.open = true
	db.File = dbfile
	return
}

// Close closing database
func (db *DB) Close() {
	db.open = false
	db.db.Close()
}

// Find returns model from database
// 		m := Model{}
// 		db.Find([]string{"bucket"}, &m)
func (db *DB) Find(path []string, id string, i interface{}) error {
	l := logit(db.Log, "FIND", path, id, nil)
	err := db.find(path, id, i)
	return l.done(err)
}

func (db *DB) find(path []string, id string, i interface{}) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.View(func(tx *bolt.Tx) error {
		var err error
		b := getBucket(tx, path)
		if b == nil {
			return errors.New("Bucket not found")
		}

		k := []byte(id)
		if err = unmarshal(b.Get(k), i); err != nil {
			return err
		}
		return nil
	})
}

// GET returns value by key
// 		val, err := db.FindValue([]string{"bucket"}, "1")
func (db *DB) Get(path []string, key string) ([]byte, error) {
	l := logit(db.Log, "GET", path, key, nil)
	v, err := db.get(path, key)
	return v, l.done(err)
}

func (db *DB) get(path []string, key string) ([]byte, error) {
	if err := db.check(path); err != nil {
		return nil, err
	}

	var v []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return errors.New("Bucket not found")
		}
		v = b.Get([]byte(key))
		return nil
	})
	return v, err
}

// Save saves model into database
// 		m := Model{Name: "Model Name"}
// 		db.Save([]string{"bucket"}, &m)
func (db *DB) Save(path []string, m mod) error {
	l := logit(db.Log, "SAVE", path, "", m)
	err := db.save(path, m)
	return l.done(err)
}

func (db *DB) save(path []string, m mod) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.Update(func(tx *bolt.Tx) error {
		b, err := createBucket(tx, path)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		id, newItem := checkID(b, m)

		enc, err := marshal(m)
		if err != nil {
			return fmt.Errorf("could not encode %s: %s", id, err)
		}

		if err := b.Put([]byte(id), enc); err != nil {
			return err
		}

		if newItem {
			addEvent("Created", m)
		} else {
			addEvent("Updated", m)
		}
		return nil
	})
}

// SaveValue saves key/value pair into database
func (db *DB) SaveValue(path []string, id string, val []byte) error {
	l := logit(db.Log, "SAVE-VALUE", path, "", val)
	err := db.saveValue(path, id, val)
	return l.done(err)
}

func (db *DB) saveValue(path []string, id string, val []byte) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.Update(func(tx *bolt.Tx) error {
		b, err := createBucket(tx, path)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return b.Put([]byte(id), val)
	})
}

// Delete deletes model from database
// 		m := Model{}
// 		db.Find([]string{"bucket"}, &m)
// 		db.Delete([]string{"bucket"}, &m)
func (db *DB) Delete(path []string, m mod) error {
	err := db.DeleteKeys(path, []string{m.GetID()})
	if err == nil {
		addEvent("Deleted", m)
	}
	return err
}

// DeleteKeys deletes records from database by keys
// 		db.DeleteKeys([]string{"bucket"}, []string{"1","2","3"})
func (db *DB) DeleteKeys(path []string, keys []string) error {
	l := logit(db.Log, "Delete", path, "", keys)
	err := db.deleteKeys(path, keys)
	return l.done(err)
}

func (db *DB) deleteKeys(path []string, keys []string) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.Update(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}
		for _, v := range keys {
			b.Delete([]byte(v))
		}
		return nil
	})
}

// DeleteBuckets deletes records from database by keys
// 		db.DeleteBuckets([]string{"bucket"}, []string{"1","2","3"})
func (db *DB) DeleteBuckets(path []string, keys []string) error {
	l := logit(db.Log, "DELETE BUCKET", path, "", keys)
	err := db.deleteBuckets(path, keys)
	return l.done(err)
}

func (db *DB) deleteBuckets(path []string, keys []string) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.Update(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}
		for _, v := range keys {
			if err := b.DeleteBucket([]byte(v)); err != nil {
				return err
			}
		}
		return nil
	})
}

// List fills models slice with records from database
// 		m := []Model{}
// 		db.List([]string{"bucket"}, &m)
// load with params
// 		db.List([]string{"bucket"}, &m, Params{Offset: 10, Limit: 30})
func (db *DB) List(path []string, dest interface{}, params ...Params) error {
	l := logit(db.Log, "LIST", path, "", params)
	err := db.list(path, dest, params...)
	return l.done(err)
}

func (db *DB) list(path []string, dest interface{}, params ...Params) error {
	if err := db.check(path); err != nil {
		return err
	}

	opts := parseParams(params)

	return db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		v := reflect.ValueOf(dest)
		if v.Kind() != reflect.Ptr {
			return errors.New("expected pointer but value passed")
		}
		if v.IsNil() {
			return errors.New("nil pointer passed")
		}

		d := reflect.Indirect(v)
		slice, err := baseType(v.Type(), reflect.Slice)
		if err != nil {
			return err
		}

		ptr := slice.Elem().Kind() == reflect.Ptr
		tp := deref(slice.Elem())

		c := b.Cursor()
		i := 0
		for k, v := cursorStart(c, opts.Reverse); k != nil; k, v = cursorNext(c, opts.Reverse) {
			if v == nil || i < opts.Offset || i > opts.Offset+opts.Limit {
				continue
			}
			i++
			item := reflect.New(tp)
			if err := unmarshal(v, item.Interface()); err != nil && err.Error() != "unexpected end of JSON input" {
				return err
			}

			if ptr {
				d.Set(reflect.Append(d, item))
			} else {
				d.Set(reflect.Append(d, reflect.Indirect(item)))
			}
		}
		return nil
	})
}

// ListKeys fills models slice with records by keys provided
// 		m := []Model{}
// 		db.ListKeys([]string{"bucket"}, [][]byte{[]byte("1"),[]byte("2")}, &m)
func (db *DB) ListKeys(path []string, keys [][]byte, dest interface{}) error {
	l := logit(db.Log, "LISTKEYS", path, "", nil)
	err := db.listKeys(path, keys, dest)
	return l.done(err)
}

func (db *DB) listKeys(path []string, keys [][]byte, dest interface{}) error {
	if err := db.check(path); err != nil {
		return err
	}

	return db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		v := reflect.ValueOf(dest)
		if v.Kind() != reflect.Ptr {
			return errors.New("expected pointer but value passed")
		}
		if v.IsNil() {
			return errors.New("nil pointer passed")
		}

		d := reflect.Indirect(v)
		slice, err := baseType(v.Type(), reflect.Slice)
		if err != nil {
			return err
		}

		ptr := slice.Elem().Kind() == reflect.Ptr
		tp := deref(slice.Elem())

		for _, key := range keys {
			v := b.Get(key)
			if v == nil {
				continue
			}
			item := reflect.New(tp)
			if err := unmarshal(v, item.Interface()); err != nil && err.Error() != "unexpected end of JSON input" {
				return err
			}

			if ptr {
				d.Set(reflect.Append(d, item))
			} else {
				d.Set(reflect.Append(d, reflect.Indirect(item)))
			}
		}
		return nil
	})
}

// ListItems returns raw records from database
func (db *DB) ListItems(path []string, params ...Params) (map[string][]byte, error) {
	l := logit(db.Log, "LIST", path, "", params)
	res := make(map[string][]byte)
	err := db.listItems(path, res, params...)
	return res, l.done(err)
}

func (db *DB) listItems(path []string, res map[string][]byte, params ...Params) error {
	if err := db.check(path); err != nil {
		return err
	}

	opts := parseParams(params)

	err := db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		i := 0
		c := b.Cursor()
		for k, v := cursorStart(c, opts.Reverse); k != nil; k, v = cursorNext(c, opts.Reverse) {
			if i < opts.Offset || i > opts.Offset+opts.Limit {
				continue
			}
			i++
			res[string(k)] = v
		}

		return nil
	})
	return err
}

// Values returns values from bucket
func (db *DB) Values(path []string, params ...Params) ([][]byte, error) {
	l := logit(db.Log, "VALUES", path, "", params)
	res, err := db.values(path, params...)
	return res, l.done(err)
}

func (db *DB) values(path []string, params ...Params) (res [][]byte, err error) {
	if err = db.check(path); err != nil {
		return
	}

	opts := parseParams(params)

	err = db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		i := 0
		c := b.Cursor()
		for k, v := cursorStart(c, opts.Reverse); k != nil; k, v = cursorNext(c, opts.Reverse) {
			if i < opts.Offset || i > opts.Offset+opts.Limit {
				continue
			}
			i++
			res = append(res, v)
		}

		return nil
	})
	return
}

// Count returns number of records in bucket
func (db *DB) Count(path []string) int {
	res := 0

	if err := db.check(path); err != nil {
		return res
	}

	db.db.View(func(tx *bolt.Tx) error {
		b := getBucket(tx, path)
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}
		res = b.Stats().KeyN
		return nil
	})
	return res
}

func (db *DB) check(path []string) error {
	if !db.open {
		return fmt.Errorf("db is not opened")
	}
	if len(path) == 0 {
		return errors.New("No bucket provided")
	}
	return nil
}

func parseParams(p []Params) Params {
	r := Params{
		Offset: 0,
		Limit:  1000,
	}

	if len(p) > 0 {
		r = p[0]
	}
	return r
}

package borm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

func createBucket(tx *bolt.Tx, buckets []string) (b *bolt.Bucket, err error) {
	b, err = tx.CreateBucketIfNotExists([]byte(buckets[0]))
	if err != nil {
		return
	}
	if len(buckets) > 1 {
		for _, v := range buckets[1:] {
			b, err = b.CreateBucketIfNotExists([]byte(v))
			if err != nil {
				return
			}
		}
	}
	return
}

func getBucket(tx *bolt.Tx, buckets []string) (b *bolt.Bucket) {
	b = tx.Bucket([]byte(buckets[0]))
	if b != nil {
		if len(buckets) > 1 {
			for _, v := range buckets[1:] {
				if b == nil {
					return nil
				}
				b = b.Bucket([]byte(v))
			}
		}
	}
	return
}

func unmarshal(data []byte, i interface{}) error {
	err := json.Unmarshal(data, i)
	if err != nil {
		return err
	}
	return nil
}

func marshal(i interface{}) ([]byte, error) {
	enc, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func nextID(b *bolt.Bucket) string {
	id, _ := b.NextSequence()
	return fmt.Sprint(id)
}

// deref is Indirect for reflect.Types
func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func baseType(t reflect.Type, expected reflect.Kind) (reflect.Type, error) {
	t = deref(t)
	if t.Kind() != expected {
		return nil, fmt.Errorf("expected %s but got %s", expected, t.Kind())
	}
	return t, nil
}

func checkID(b *bolt.Bucket, m mod) (id string, newItem bool) {
	id = m.GetID()
	if id == "" {
		newItem = true
		id = nextID(b)
		m.setID(id)
		if m1, ok := m.(modCreate); ok {
			m1.setCreation()
		}
		if m1, ok := m.(modUpdate); ok {
			m1.touchModel()
		}
	} else {
		if m1, ok := m.(modUpdate); ok {
			m1.touchModel()
		}
	}
	return
}

func typeName(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

var addEvent = func(name string, m mod) {
	go Events.Pub(eventName(name, m), m)
}

func eventName(name string, m mod) string {
	return typeName(m) + name
}

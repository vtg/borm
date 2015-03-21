borm
======

boltdb tiny orm for processing model structures.

######Model structure:
```go
type Person struct {
  // required
  borm.Model
  
  // Model fields
  Name    string
  Age     int

  // optional. used to add Created field into model. Will be set automaticaly on creation
  borm.CreateTime
  
  // optional. used to add Updated field into model. Will be updated automaticaly on each model save
  borm.UpdateTime
}
```

######Usage example
```go
package main

import (
	"fmt"
	"log"

	"github.com/vtg/borm"
)

type Person struct {
	borm.Model

	Name string
	Age  int

	borm.CreateTime
	borm.UpdateTime
}

func main() {
	db, err := borm.Open("database.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Log = true
	defer db.Close()

	// the bucket that will store Person
	bucket := []string{"people"}

	// creating Person
	p := Person{Name: "John Doe"}
	db.Save(bucket, &p)
	fmt.Println(p.ID, p.Name, p.Age)

	//updating Person
	p.Age = 10
	db.Save(bucket, &p)
	fmt.Println(p.ID, p.Name, p.Age)

	// get Person from database
	p1 := Person{}
	db.Find(bucket, p.ID, &p1)
	fmt.Println(p1.ID, p1.Name, p1.Age)

	// list people
	people := []Person{}
	db.List(bucket, &people)
	fmt.Println(people)

	// delete Person from database
	db.Delete(bucket, &p)

	// count peoples
	fmt.Println(db.Count(bucket))
}
```

######Events
borm has events subscription support.  
There are 3 types of Events "Created", "Updated" and "Deleted".  
Event names prefixed with model type name.
```go
//Subscribing to Person creation event
borm.Events.Sub("PersonCreated", func(e *pubsub.Event) {
	fmt.Println(e.Name, e.Objects[0].(*Person))
})
```

GoDoc https://godoc.org/github.com/vtg/borm

#####Author

VTG - http://github.com/vtg

##### License

Released under the [MIT License](http://www.opensource.org/licenses/MIT).

[![GoDoc](https://godoc.org/github.com/vtg/borm?status.png)](http://godoc.org/github.com/vtg/borm)

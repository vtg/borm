package borm

import "time"

type mod interface {
	GetID() string
	setID(id string)
}

type modCreate interface {
	setCreation()
}

type modUpdate interface {
	touchModel()
}

// Model
type Model struct {
	MID
	validator
}

// MID
type MID struct {
	ID string
}

// GetID returns id of model
func (i *MID) GetID() string {
	return i.ID
}

func (i *MID) setID(id string) {
	i.ID = id
}

// CreateTime struct for managing creation time
type CreateTime struct {
	Created time.Time
}

func (c *CreateTime) setCreation() {
	c.Created = time.Now()
}

// UpdateTime struct for managing update time
type UpdateTime struct {
	Updated time.Time
}

func (u *UpdateTime) touchModel() {
	u.Updated = time.Now()
}

package borm

import (
	"strconv"
	"time"
)

type mod interface {
	GetID() string
	setID(id string)
	Created() time.Time
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

// Created returns creation time
func (i *MID) Created() time.Time {
	u, _ := strconv.ParseInt(i.ID, 10, 64)
	return time.Unix(0, u)
}

func (i *MID) setID(id string) {
	i.ID = id
}

// UpdateTime struct for managing update time
type UpdateTime struct {
	Updated time.Time
}

func (u *UpdateTime) touchModel() {
	u.Updated = time.Now()
}

package calcifer

import "time"

// A Model is a Go-native representation of a document that can be stored in Firestore.
// Embeded `Model` into your own struct to define other types of models.
type Model struct {
	ID         string
	CreateTime time.Time
	UpdateTime time.Time
}

// The model interface is implemented only by calcifer.Model and structs that embed it.
type model interface {
	setID(string)
	setCreateTime(time.Time)
	setUpdateTime(time.Time)
}

func (m *Model) setID(id string) {
	m.ID = id
}

func (m *Model) setCreateTime(t time.Time) {
	m.CreateTime = t
}

func (m *Model) setUpdateTime(t time.Time) {
	m.UpdateTime = t
}

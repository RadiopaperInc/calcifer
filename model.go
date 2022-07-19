// Copyright 2022 Radiopaper Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package calcifer

import "time"

// A Model is a Go-native representation of a document that can be stored in Firestore.
// Embeded `Model` into your own struct to define other types of models.
type Model struct {
	ID         string    `calcifer:"id"`
	CreateTime time.Time `calcifer:"create_time"`
	UpdateTime time.Time `calcifer:"update_time"`
}

// // The MutableModel interface is satisfied only by pointers to calcifer.Model and structs that embed it.
// type MutableModel interface {
// 	setID(string)
// 	setCreateTime(time.Time)
// 	setUpdateTime(time.Time)
// }

func (m Model) setID(id string) {
	m.ID = id
}

func (m Model) setCreateTime(t time.Time) {
	m.CreateTime = t
}

func (m Model) setUpdateTime(t time.Time) {
	m.UpdateTime = t
}

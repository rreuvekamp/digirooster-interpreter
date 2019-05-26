// Copyright Remi Reuvekamp 2018-2019

//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package drparser

import (
	"encoding/json"
	"io"
	"strings"
	"time"
)

// Wrapper is the structure of the raw JSON from the digirooster endpoint.
// JSON unmarshalling is handled in two stages, wrapper wraps the real data.
//
// Types wrapper, data and activity are the format of the input JSON.
type Wrapper struct {
	D string
}
type Data struct {
	ScheduleStart int
	ScheduleEnd   int
	ScheduleId    string
	ChangeData    []map[string]interface{}
	ActivityData  []Activity
}
type Activity struct {
	ID          string
	Description string
	Location    string
	Student     string
	Staff       string
	StartUTS    int64 `json:"start"`
	EndUTS      int64 `json:"end"`
	Week        int
	Width       int
	Left        int
}

func ParseJSON(r io.Reader) (Data, error) {
	w := Wrapper{}

	dec := json.NewDecoder(r)
	err := dec.Decode(&w)
	if err != nil {
		return Data{}, err
	}

	d := Data{}

	err = json.Unmarshal([]byte(w.D), &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

type NewData struct {
	Name     string
	Start    string
	End      string
	Teachers []struct {
		Code string
	}
	Rooms []struct {
		Name string
	}
	Subgroups []struct {
		Name string
	}
}

func (d NewData) rooms() string {
	loc := []string{}
	for _, r := range d.Rooms {
		loc = append(loc, r.Name)
	}

	return strings.Join(loc, ",")
}

func (d NewData) groups() string {
	loc := []string{}
	for _, r := range d.Subgroups {
		loc = append(loc, r.Name)
	}

	return strings.Join(loc, ",")
}

func (d NewData) staff() string {
	loc := []string{}
	for _, r := range d.Teachers {
		loc = append(loc, r.Code)
	}

	return strings.Join(loc, ",")
}

func ParseJSONNew(r io.Reader) (Data, error) {
	d := []NewData{}

	dec := json.NewDecoder(r)
	err := dec.Decode(&d)
	if err != nil {
		return Data{}, err
	}

	return newToOld(d), nil
}

func newToOld(newData []NewData) Data {
	acts := make([]Activity, 0, len(newData))

	for _, d := range newData {
		start, _ := time.Parse("2006-01-02T15:04:05", d.Start)
		end, _ := time.Parse("2006-01-02T15:04:05", d.End)

		act := Activity{
			Description: d.Name,
			Location:    d.rooms(),
			Student:     d.groups(),
			Staff:       d.staff(),
			StartUTS:    start.Unix() * 1000,
			EndUTS:      end.Unix() * 1000,
		}

		acts = append(acts, act)
	}

	return Data{
		ActivityData: acts,
	}
}

func (a Activity) Start() time.Time {
	return time.Unix(int64(a.StartUTS/1000), 0)
}

func (a Activity) End() time.Time {
	return time.Unix(int64(a.EndUTS/1000), 0)
}

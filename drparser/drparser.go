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
	Start       int
	End         int
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

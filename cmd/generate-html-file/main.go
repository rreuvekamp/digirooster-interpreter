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

package main

import (
	"crypto/md5"
	"fmt"

	"html/template"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.remi.im/remi/digirooster-interpreter/drparser"
)

type templateData struct {
	ClassName        string
	GeneratedTimeStr string
	Weeks            []templateWeek
}
type templateWeek struct {
	WeekNumber uint8
	Hours      int
	StartAt    int
	EndAt      int
	Days       []templateDay
}
type templateDay struct {
	DayString  string
	Hours      int
	Activities []templateActivity
}
type templateActivity struct {
	Desc     string
	Course   string
	OrigDesc string
	Loc      string
	Staff    []templateStaff
	Classes  []string

	Start    time.Time
	End      time.Time
	StartStr string
	EndStr   string

	CourseColour string

	Padding int
	Height  int

	NonImportant bool
	Important    bool
}
type templateStaff struct {
	ID   string
	Name string
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: digirooster-interpeter <json-file.json> <class-name>")
		os.Exit(1)
	}

	d, err := readData(os.Args[1])
	if err != nil {
		fmt.Println("Could not open provided JSON file. Does it exist and is it valid JSON?")
		fmt.Println("Error:")
		log.Fatal(err)
		os.Exit(2)
	}

	tmpl, err := template.ParseFiles("page.tmpl")
	if err != nil {
		fmt.Println("Could not parse template:")
		log.Fatal(err)
		os.Exit(3)
	}

	tw, err := toTemplateWeeks(d)

	err = tmpl.Execute(os.Stdout, templateData{
		os.Args[2],
		time.Now().Format("2006-01-02 15:04"),
		tw,
	})
}

func readData(fn string) (drparser.Data, error) {
	f, err := os.Open(fn)
	if err != nil {
		return drparser.Data{}, err
	}
	defer f.Close()

	return drparser.ParseJSONNew(f)
}

func toTemplateWeeks(d drparser.Data) ([]templateWeek, error) {
	weeks := make(map[int]templateWeek)

	// Transform the data into separate weeks with days.
	for _, a := range d.ActivityData {
		start := a.Start()
		end := a.End()

		year, weekNumber := start.ISOWeek()

		key, _ := strconv.Atoi(fmt.Sprintf("%.4d%.2d", year, weekNumber))

		if _, ok := weeks[key]; !ok {
			weeks[key] = templateWeek{
				WeekNumber: uint8(weekNumber),
				Days:       make([]templateDay, 5, 5),
			}
		}

		course, niceName := descName(a.Description)

		ta := templateActivity{
			niceName,
			course,
			a.Description,
			a.Location,
			staffName(a.Staff),
			classNames(a.Student),
			start,
			end,
			start.Format("15:04"),
			end.Format("15:04"),
			hashColourCode(course),
			0,
			0,
			isNonImportant(a),
			isImportant(a),
		}

		if weeks[key].Days[start.Weekday()-1].DayString == "" {
			weeks[key].Days[start.Weekday()-1].DayString = start.Format("Mon 2 Jan")
		}

		weeks[key].Days[start.Weekday()-1].Activities = append(weeks[key].Days[start.Weekday()-1].Activities, ta)
	}

	// Loop over the created weeks one more time.
	// Calculate the activity height, the vertical distance between
	// activities and the amount of hours per day.
	for k, w := range weeks {
		// Determine the times of the earlier start, and last ending, activity of this week.
		first := 0
		last := 0
		for _, d := range w.Days {
			for _, a := range d.Activities {
				start, _ := strconv.Atoi(a.Start.Format("1504"))
				end, _ := strconv.Atoi(a.End.Format("1504"))

				if first == 0 || start < first {
					first = start
				}
				if last == 0 || end > last {
					last = end
				}
			}
		}

		wMinutes := 0
		for di, d := range w.Days {
			dMinutes := 0
			cur := first
			for ai, a := range d.Activities {
				start, _ := strconv.Atoi(a.Start.Format("1504"))
				end, _ := strconv.Atoi(a.End.Format("1504"))

				if !a.NonImportant {
					dMinutes += int(a.End.Sub(a.Start).Minutes())
				}

				weeks[k].Days[di].Activities[ai].Height = (end - start) / 2

				weeks[k].Days[di].Activities[ai].Padding = (start - cur) / 2

				cur = end
			}

			wMinutes += dMinutes

			weeks[k].Days[di].Hours = dMinutes / 60
		}

		w.Hours = wMinutes / 60
		w.StartAt = first
		w.EndAt = last
		weeks[k] = w
	}

	// Sort the days, as the might be scrambled.
	var keys []int
	for k := range weeks {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	weeksSlice := make([]templateWeek, len(keys), len(keys))
	for i, k := range keys {
		weeksSlice[i] = weeks[k]
	}

	return weeksSlice, nil
}

// isNonImportant determines if an activity should be marked as unimportant.
func isNonImportant(a drparser.Activity) bool {
	desc := strings.ToLower(a.Description)
	if strings.Contains(desc, "honours") || strings.Contains(desc, "panel gesprek") {
		return true
	}

	return false
}

// isImportant determines if an activity should be marked as important.
func isImportant(a drparser.Activity) bool {
	desc := strings.ToLower(a.Description)
	if strings.Contains(desc, "inzage") || strings.Contains(desc, "toets") || strings.Contains(desc, "tentamen") || strings.Contains(desc, "presentatie") {
		return true
	}

	return false
}

// staffName gives the staff name depending on the given abbreviation.
func staffName(short string) []templateStaff {
	split := strings.Split(short, ", ")

	sort.Strings(split)

	names := []templateStaff{}

	for _, sh := range split {
		n := ""
		s := strings.ToUpper(sh)
		switch s {
		case "HATJ":
			n = "T. Harkema"
		case "BRUM":
			n = "M. de Bruin"
		case "BREJ":
			n = "J. Bredek"
		case "HEBL":
			n = "B. Heijne"
		case "BROH":
			n = "H. Brouwers"
		case "BIKO":
			n = "K. Bijker"
		case "NOLI":
			n = "L. Noordhuis"
		case "KEHT":
			n = "T. van Keulen"
		case "THAR":
			n = "A. Thuss"
		case "NIEV":
			n = "E. Nijkamp"
		case "BJAB":
			n = "J. de Boer"
		case "HOEM":
			n = "M. Hoebe"
		case "KOFA":
			n = "F. de Kooi"
		case "STRI":
			n = "I. Stroeffe"
		case "RIHH":
			n = "H. Rietdijk"
		case "BABA":
			n = "B. Barnard"
		}

		staff := templateStaff{
			ID: sh,
		}

		if len(n) > 0 {
			staff.Name = n
		}

		names = append(names, staff)

	}

	return names
}

// classNames formats the comma separated list of class names and
// transforms it to an array.
func classNames(orig string) []string {
	split := strings.Split(orig, ", ")

	classes := make([]string, len(split))
	for i, s := range split {
		// Classes have a prefix like "BF\"
		split2 := strings.Split(s, "\\")
		if len(split2) > 1 {
			s = split2[len(split2)-1] // last element op split2
		}
		s = strings.Replace(s, "ITV", "ITV-", 1)
		classes[i] = s
	}

	sort.Strings(classes)

	return classes
}

// descName formats the activity description.
func descName(orig string) (course string, name string) {
	items := strings.Split(orig, "/")
	if len(items) > 1 {
		course = items[0]

		nameItems := make([]string, 0, len(items)-1)
		for _, item := range items[1:] {
			_, err := strconv.Atoi(item)
			if err == nil {
				// Skip numeric name parts.
				continue
			}
			nameItems = append(nameItems, item)
		}
		name = strings.Join(nameItems, "/")
	} else {
		course = ""
		name = orig
	}

	split := strings.Split(name, " ")
	for i, s := range split {
		sl := strings.ToLower(s)
		switch sl {
		case "pr", "wc":
			s = "practicum"
		case "th":
			s = "lecture"
		}
		split[i] = s
	}

	name = strings.Join(split, " ")

	return
}

func hashColourCode(str string) string {
	if len(str) == 0 {
		return "ffffff"
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return string(hash[:6])
}

package main

import "os"
import "encoding/json"
import "strings"
import "log"
import "fmt"
import "html/template"
import "time"
import "sort"
import "strconv"

type wrapper struct {
	D string
}

type data struct {
	ScheduleStart int
	ScheduleEnd int
	ScheduleId string
	ChangeData []map[string]interface{}
	ActivityData []activity
}

type activity struct {
	ID string
	Description string
	Location string
	Student string
	Start int
	End int
	Week int
	Width int
	Left int
}


type templateData struct {
	ClassName string
	GeneratedTimeStr string
	Weeks []templateWeek
}

type templateWeek struct {
	WeekNumber uint8
	Hours int
	StartAt int
	EndAt int
	Days []templateDay
}

type templateDay struct {
	DayString string
	Hours int
	Activities []templateActivity
}

type templateActivity struct {
	Desc string
	Loc string
	Classes []string

	Start time.Time
	End time.Time
	StartStr string
	EndStr string

	Padding int
	Height int

	NonImportant bool
	Important bool
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: digirooster-interpeter <json-file.json> <class-name>")
		os.Exit(1)
	}

	d, err := parseJSON(os.Args[1])
	if err != nil {
		fmt.Println("Could not open provided JSON file. Does it exist and is it valid JSON?")
		fmt.Println("Error:")
		log.Fatal(err)
		os.Exit(2)
	}

	//for _,a := range d.ActivityData {
		//fmt.Println(a.Description, a.Location, a.Student)
	//}

	tmpl, err := template.ParseFiles("page.tmpl")
	if err != nil {
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

func toTemplateWeeks(d data) ([]templateWeek, error) {
	weeks := make(map[int]templateWeek)

	for _, a := range d.ActivityData {
		start := time.Unix(int64(a.Start/1000), 0)
		end := time.Unix(int64(a.End/1000), 0)

		year, weekNumber := start.ISOWeek()

		key, _ := strconv.Atoi(fmt.Sprintf("%.4d%.2d", year, weekNumber))

		if _, ok := weeks[key]; !ok {
			weeks[key] = templateWeek{
				WeekNumber: uint8(weekNumber),
				Days: make([]templateDay, 5, 5),
			}
		}

		split := strings.Split(a.Student, ", ")

		classes := make([]string, len(split))
		for i, s := range split {
			// Classes have a prefix like "BF\"
			split2 := strings.Split(s, "\\")
			classes[i] = split2[len(split2)-1] // last element op split2
		}

		ta := templateActivity {
			a.Description,
			a.Location,
			classes,
			start,
			end,
			start.Format("15:04"),
			end.Format("15:04"),
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

	for k, w := range weeks {
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

				weeks[k].Days[di].Activities[ai].Height = (end-start)/2

				weeks[k].Days[di].Activities[ai].Padding = (start-cur)/2

				cur = end
			}

			wMinutes += dMinutes

			weeks[k].Days[di].Hours = dMinutes/60
		}

		w.Hours = wMinutes/60
		w.StartAt = first
		w.EndAt = last
		weeks[k] = w
	}

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

func parseJSON(fileName string) (data, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return data{}, err
	}
	defer f.Close()

	w := wrapper{}

	dec := json.NewDecoder(f)
	err = dec.Decode(&w)
	if err != nil {
		return data{}, err
	}

	d := data{}

	err = json.Unmarshal([]byte(w.D), &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

func isNonImportant(a activity) bool {
	desc := strings.ToLower(a.Description)
	if strings.Contains(desc, "honours") || strings.Contains(desc, "panel gesprek") {
		return true
	}

	return false
}

func isImportant(a activity) bool {
	desc := strings.ToLower(a.Description)
	if strings.Contains(desc, "inzage") || strings.Contains(desc, "toets") || strings.Contains(desc, "tentamen") || strings.Contains(desc, "presentatie") {
		return true
	}

	return false
}
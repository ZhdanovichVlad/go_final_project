package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ZhdanovichVlad/go_final_project/api"
	http_server "github.com/ZhdanovichVlad/go_final_project/http-server"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Constants for defining the database search criterion
const (
	FullOutput = iota
	DateSearch
	TextSearch
)

// a structure for receiving tasks from the frontend and sending them back to the frontend
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// Structure for returning a JSON-formatted response when adding a task
type TaskResponseID struct {
	ID string `json:"id"`
}

// ApiNextDate function for the api.NextDate handle. Sends the user the date when the next task should be performed.
func ApiNextDate(w http.ResponseWriter, req *http.Request) {

	now := req.FormValue("now")
	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		log.Panic(err)
	}
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")
	answear, err := api.NextDate(nowTime, date, repeat)
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write([]byte(answear))
	if err != nil {
		log.Printf("error while writing response: %v", err)
	}

}

// PostTask. Нandler function that adds tasks to the database
func PostTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http_server.ResponseJson("Error when reading from req.Body", http.StatusBadRequest, err, w)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		if task.Title == "" {
			http_server.ResponseJson("the Title field cannot be empty", http.StatusBadRequest, nil, w)
			return
		}
		var date string // - create a date that will be written to the database
		// check for date value and repeat rules.
		// Depending on the rules set, set the date to be written to the database.
		switch {
		case task.Date == time.Now().Format("20060102"):
			date = time.Now().Format("20060102")
		case task.Repeat == "" && task.Date == "":
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date == "":
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date != "":
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
				return
			}
			if dataTime.After(time.Now()) {
				date = task.Date
			} else {
				date, err = api.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http_server.ResponseJson("api.NextDate Error", http.StatusBadRequest, err, w)
					return
				}
			}
		case task.Repeat == "" && task.Date != "":
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
				return
			}
			if dataTime.Before(time.Now()) {
				date = time.Now().Format("20060102")
			} else {
				date = task.Date
			}
		default:
			date = task.Date
		}

		//// add a new task to the database
		lastID, err := serverJob.AddTask(date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		var answear TaskResponseID = TaskResponseID{lastID}

		answearJSON, err := json.Marshal(&answear)
		if err != nil {
			http_server.ResponseJson("Deserialization error", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(answearJSON)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

// GetTasks function to get the nearest tasks. The number of tasks is set by the variable NumberOfOupTasksString. The variable is defined in the .env file
// this function also handles search by date or text.
func GetTasks(serverJob ServerJob, NumberOfOutTasksString string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		NumberOfOutTasksInt, err := strconv.Atoi(NumberOfOutTasksString)
		if err != nil {
			http_server.ResponseJson("NumberOfOutTasksString to NumberOfOutTasksInt convert error. Сheck the .env file.", http.StatusInternalServerError, err, w)
			return
		}
		reqQuery := req.URL.Query().Get("search")
		// Based on the reqQuery values, we have three possible cases. Each will have its own encoding
		//1-case. reqQuery is empty.         Code 0. The NumberOfOutTasksString of the nearest tasks specified in the NumberOfOutTasksString parameter is returned
		//2-case. reqQuery contains date -   Code 1. The database is searched by date.
		//3-case. reqQuery contains text -   Code 2. The database is searched by title and comment fields
		var code int
		_, err = time.Parse("02.01.2006", reqQuery)
		if reqQuery == "" {
			code = FullOutput
		} else if err == nil {
			code = DateSearch
		} else {
			code = TextSearch
		}

		var answearRows []Task  // creating a future reply to the user
		if code == FullOutput { // case when returning the nearest tasks
			answearRows, err = serverJob.GetTasks(NumberOfOutTasksInt)
			if err != nil {
				http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
				return
			}
		} else { // case of searching by date or text
			answearRows, err = serverJob.SearchTasks(code, reqQuery, NumberOfOutTasksInt)
			if err != nil {
				http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
				return
			}
		}

		if answearRows == nil {
			answearRows = make([]Task, 0)
		}
		var answearsStruck = map[string][]Task{"tasks": answearRows}
		jsonMsg, err := json.Marshal(answearsStruck)
		if err != nil {
			http_server.ResponseJson("error when using the Marshal function", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

// GetTask function returns a task from the database by the specified ID in the Query
func GetTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		qeryID := req.URL.Query().Get("id")
		if qeryID == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		answerRow, err := serverJob.GetTask(qeryID)
		if err != nil {
			http_server.ResponseJson("error when retrieving tasks from the database", http.StatusInternalServerError, err, w)
			return
		}
		jsonMsg, err := json.Marshal(answerRow)
		if err != nil {
			http_server.ResponseJson("Error when using the Marshal function", http.StatusInternalServerError, err, w)
			return

		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

// CorrectTask  function that CorrectTask in the database
func CorrectTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http_server.ResponseJson("Error when reading from req.Body", http.StatusBadRequest, err, w)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http_server.ResponseJson("deserialization error", http.StatusInternalServerError, err, w)
			return
		}

		_, err = strconv.Atoi(task.ID) // проверка, что в поле ID передана цифра
		if err != nil {
			http_server.ResponseJson("wrong ID", http.StatusBadRequest, err, w)
			return

		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			http_server.ResponseJson("Parse to date error", http.StatusBadRequest, err, w)
			return
		}

		if task.Title == "" { // the Title value must always contain a description of the task. If it is empty, we return an error
			http_server.ResponseJson("The title is empty", http.StatusBadRequest, nil, w)
			return

		} // the date of the task must be greater than today. Otherwise we return an error
		if dateTime.Before(time.Now()) && dateTime.Equal(time.Now()) {
			http_server.ResponseJson("the date can't be less than today", http.StatusBadRequest, nil, w)
			return
		}

		// The Repeat rule can be an empty string or begin with the following characters y,d,m,w.
		// If Repeat specifies the other, we return an error.
		// We immediately check if the string is empty. If it is not, we do
		// Convert the string to an array of runes and compare the first rune.

		if task.Repeat != "" {
			startChars := []rune{'y', 'd', 'm', 'w'}
			firstRuneRepeat := []rune(task.Repeat)[0]
			flagChek := true
			for _, s := range startChars {
				if s == firstRuneRepeat {
					flagChek = false
				}
			}
			if flagChek {
				http_server.ResponseJson("the rule for repetition has the wrong format.", http.StatusBadRequest, nil, w)
				return
			}
		}

		// update the task in the database
		err = serverJob.UpdateTask(task)
		if err != nil {
			http_server.ResponseJson("error when updating data on the server", http.StatusInternalServerError, err, w)
			return
		}
		jsonMsg, err := json.Marshal(Task{})
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}

	}
}

// DoneTask function to delete and update a task in the database by a given ID if it has been executed
func DoneTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		_, err := strconv.Atoi(idDone) // проверка, на то что в поле ID передано число
		if err != nil {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, err, w)
			return
		}
		task, err := serverJob.GetTask(idDone) // проверка, что задание существует в базе данных
		if task.ID == "" {
			http_server.ResponseJson("id not found in the database", http.StatusBadRequest, err, w)
			return
		}
		// If the value to repeat is not specified in the database. We delete it.
		// Otherwise (block else), we calculate a new date and update the task
		if task.Repeat == "" {
			err = serverJob.DeleteTask(idDone)
			if err != nil {
				http_server.ResponseJson("failed to delete the task", http.StatusInternalServerError, err, w)
				return
			}
		} else {
			newDateString, err := api.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http_server.ResponseJson("api.NextDate Error", http.StatusInternalServerError, err, w)
				return
			}
			err = serverJob.UpdateDateTask(idDone, newDateString)
			if err != nil {
				http_server.ResponseJson("data update error", http.StatusInternalServerError, err, w)
				return
			}
		}
		var answear = map[string]any{}
		jsonMsg, err := json.Marshal(answear)
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

// DeleteTask function to delete a task by the given Query id.
func DeleteTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, nil, w)
			return
		}
		_, err := strconv.Atoi(idDone) // проверка, передано ли в поле ID цифра
		if err != nil {
			http_server.ResponseJson("wrong id", http.StatusBadRequest, err, w)
			return
		}

		err = serverJob.DeleteTask(idDone) // Delete the task. If the deletion fails, an error is returned
		if err != nil {
			http_server.ResponseJson("failed to delete the task", http.StatusBadRequest, err, w)
			return
		}

		//generate and return empty json as a response
		var answear = map[string]any{}
		jsonMsg, err := json.Marshal(answear)
		if err != nil {
			http_server.ResponseJson("error when generating the serialization of the response", http.StatusInternalServerError, err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonMsg)
		if err != nil {
			log.Printf("error while writing response: %v", err)
		}
	}
}

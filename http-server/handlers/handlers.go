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
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title" binding:"required"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TaskResponseID struct {
	ID int64 `json:"id"`
}

type ServerJob interface {
	AddTask(date string, title string, comment string, repeat string) (int64, error)
	GetTasks(NumberOfOuptuTasks int, tasks []Task) ([]Task, error)
	GetTask(id string) (Task, error)
	UpdateTask(task Task) error
	DeleteTask(idTask string) error
	UpdateDateTask(idTask string, newDateString string) error
	SearchTasks(code int, searchQuery string, NumberOfOuptuTasks int) ([]Task, error)
}

// ApiNextDate. Handler for api.NextDate
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
	w.Write([]byte(answear))

}

// PostTask. Нandler function that adds tasks to the database
func PostTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Error when reading from req.Body"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if task.Title == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		var date string
		switch {
		case task.Date == time.Now().Format("20060102"):
			date = time.Now().Format("20060102")
		case task.Repeat == "" && task.Date == "": // ОБА ПУСТЫЕ
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date == "": // дата пустая
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date != "": // ОБА НЕ ПУСТЫЕ
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Parse to date error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			if dataTime.After(time.Now()) {
				date = task.Date
			} else {
				date, err = api.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"api.NextDate Error"}, false)
					w.WriteHeader(errInt)
					w.Header().Set("Content-Type", "application/json; charset=UTF-8")
					w.Write(msg)
					return
				}
			}
		case task.Repeat == "" && task.Date != "": // ДАТА НЕ ПУСТАЯ
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Parse to date error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			if dataTime.Before(time.Now()) {
				date = time.Now().Format("20060102")
			} else {
				date = task.Date
			}
		default:
			date = task.Date
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"api.NextDate Error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}
		dataTime, err := time.Parse("20060102", date)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Parse to date error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		lastID, err := serverJob.AddTask(dataTime.Format("20060102"), task.Title, task.Comment, task.Repeat)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		} else {
			var answear TaskResponseID = TaskResponseID{lastID}
			answearJSON, _ := json.Marshal(&answear)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(answearJSON)
			return
		}
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

	}
}

// GetTasksHundler. function to get the nearest tasks. The number of tasks is set by the variable NumberOfOuptuTasksString
func GetTasksHundler(serverJob ServerJob, NumberOfOuptuTasksString string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		NumberOfOuptuTasksInt, err := strconv.Atoi(NumberOfOuptuTasksString)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"NumberOfOuptuTasksString to NumberOfOuptuTasksInt convert error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		reqQuery := req.URL.Query().Get("search")
		//since we have three possible cases. We will code them in the following way.
		//1-case. The queue is empty - code 0. 2-case. The queue contains a date - code 1. 3-case. The queue contains text - code 2.
		var code int
		_, err = time.Parse("02.01.2006", reqQuery)
		if reqQuery == "" {
			code = 0
		} else if err == nil {
			code = 1
		} else {
			code = 2
		}

		var answearRows []Task
		if code == 0 {
			answearRows, err = serverJob.GetTasks(NumberOfOuptuTasksInt, answearRows)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when retrieving tasks from the database"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		} else {
			answearRows, err = serverJob.SearchTasks(code, reqQuery, NumberOfOuptuTasksInt)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when retrieving tasks from the database"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}

		if answearRows == nil {
			answearRows = make([]Task, 0)
		}
		var answearsStruck = map[string][]Task{"tasks": answearRows}
		jsonMsg, err := json.Marshal(answearsStruck)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Error when using the Marshal function"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(jsonMsg)
	}
}

// GetTaskHundler. Нandler function that Get a Task from database
func GetTaskHundler(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		qeryID := req.URL.Query().Get("id")
		if qeryID == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong id"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		answearRow, err := serverJob.GetTask(qeryID)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when retrieving tasks from the database"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		jsonMsg, err := json.Marshal(answearRow)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Error when using the Marshal function"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(jsonMsg)
	}
}

// CorrectTask. Нandler function that CorrectTask in the database
func CorrectTask(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"Error when reading from req.Body"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		lenRepeat := len(strings.Split(task.Repeat, " "))
		_, err = strconv.Atoi(task.ID)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"the date is in the wrong format"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if task.Title == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"The title is empty"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if dateTime.After(time.Now()) || dateTime.Equal(time.Now()) {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"the date can't be less than today"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if lenRepeat != 2 && lenRepeat != 0 {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"the rule for repetition has the wrong format."}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		err = serverJob.UpdateTask(task)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when updating data on the server"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		msg, err := json.Marshal(Task{})
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when generating the serialization of the response"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)

	}
}

func DoneTaskHundler(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong id"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		_, err := strconv.Atoi(idDone)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		task, err := serverJob.GetTask(idDone)
		if task.ID == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"id not found in the database"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return

		}
		if task.Repeat == "" {
			err = serverJob.DeleteTask(idDone)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"failed to delete the task"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		} else {
			newDateString, err := api.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"api.NextDate Error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			err = serverJob.UpdateDateTask(idDone, newDateString)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"data update error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}
		var answear = map[string]any{}
		msg, err := json.Marshal(answear)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when generating the serialization of the response"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
	}
}

func DeleteTaskHundler(serverJob ServerJob) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idDone := req.URL.Query().Get("id")
		if idDone == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong id"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		_, err := strconv.Atoi(idDone)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		err = serverJob.DeleteTask(idDone)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"failed to delete the task"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		var answear = map[string]any{}
		msg, err := json.Marshal(answear)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when generating the serialization of the response"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
	}
}

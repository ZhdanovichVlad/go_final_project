package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ZhdanovichVlad/go_final_project/api"
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

type TaskResponseError struct {
	Message string `json:"error"`
}

type ServerJob interface {
	AddTask(date string, title string, comment string, repeat string) (int64, error)
	GetTasks(NumberOfOuptuTasks int, tasks []Task) ([]Task, error)
	GetTask(id string) (Task, error)
	UpdateTask(task Task) error
	DeleteTask(idTask string) error
	UpdateDateTask(idTask string, newDateString string) error
}

func jsonErrorMarshal(message TaskResponseError, isBadRequest bool) ([]byte, int) {
	var returnStatus int
	if isBadRequest {
		returnStatus = http.StatusBadRequest
	} else {
		returnStatus = http.StatusInternalServerError
	}
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return []byte(err.Error()), http.StatusInternalServerError
	}
	return jsonMsg, returnStatus
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
			msg, errInt := jsonErrorMarshal(TaskResponseError{"Error when reading from req.Body"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"dicerelization error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if task.Title == "" {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		var date string
		switch {
		case task.Date == time.Now().Format("20060102"):
			date = time.Now().Format("20060102")
		case task.Repeat == "" && task.Date == "":
			date = time.Now().Format("20060102")
		case task.Repeat != "" && task.Date != "":

			date, err = api.NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
			if err != nil {
				msg, errInt := jsonErrorMarshal(TaskResponseError{"api.NextDate Error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		case task.Repeat == "" && task.Date != "":
			dataTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				msg, errInt := jsonErrorMarshal(TaskResponseError{"Parse to date error"}, false)
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
				msg, errInt := jsonErrorMarshal(TaskResponseError{"api.NextDate Error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}
		dataTime, err := time.Parse("20060102", date)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"Parse to date error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		lastID, err := serverJob.AddTask(dataTime.Format("20060102"), task.Title, task.Comment, task.Repeat)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"dicerelization error"}, true)
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
			msg, errInt := jsonErrorMarshal(TaskResponseError{"dicerelization error"}, true)
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
			msg, errInt := jsonErrorMarshal(TaskResponseError{"NumberOfOuptuTasksString to NumberOfOuptuTasksInt convert error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		var answearRows []Task
		answearRows, err = serverJob.GetTasks(NumberOfOuptuTasksInt, answearRows)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"error when retrieving tasks from the database"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if answearRows == nil {
			answearRows = make([]Task, 0)
		}
		var answearsStruck = map[string][]Task{"tasks": answearRows}
		jsonMsg, err := json.Marshal(answearsStruck)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"Error when using the Marshal function"}, false)
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
			msg, errInt := jsonErrorMarshal(TaskResponseError{"wrong id"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		answearRow, err := serverJob.GetTask(qeryID)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"error when retrieving tasks from the database"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		jsonMsg, err := json.Marshal(answearRow)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"Error when using the Marshal function"}, false)
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
			msg, errInt := jsonErrorMarshal(TaskResponseError{"Error when reading from req.Body"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"dicerelization error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		lenRepeat := len(strings.Split(task.Repeat, " "))
		_, err = strconv.Atoi(task.ID)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"the date is in the wrong format"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		if task.Title == "" {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"The title is empty"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if dateTime.After(time.Now()) || dateTime.Equal(time.Now()) {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"the date can't be less than today"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		if lenRepeat != 2 && lenRepeat != 0 {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"the rule for repetition has the wrong format."}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		err = serverJob.UpdateTask(task)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"error when updating data on the server"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		msg, err := json.Marshal(Task{})
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"error when generating the serialization of the response"}, false)
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
		fmt.Println("ID до проверки", idDone)
		if idDone == "" {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"wrong id"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		_, err := strconv.Atoi(idDone)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		fmt.Println("ID для отправки на поиск", idDone)
		task, err := serverJob.GetTask(idDone)
		fmt.Println("Получили task", task)
		if task.Repeat == "" {
			fmt.Println("удаляем задачу", task)
			err = serverJob.DeleteTask(idDone)
			if err != nil {
				fmt.Println("удаление неудачно", err)
				msg, errInt := jsonErrorMarshal(TaskResponseError{"failed to delete the task"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		} else {
			fmt.Println("обновление задачи", task)
			newDateString, err := api.NextDate(time.Now(), task.Date, task.Repeat)
			fmt.Println("новая дата", newDateString)
			if err != nil {
				fmt.Println("ошибка при получении новой даты", err)
				msg, errInt := jsonErrorMarshal(TaskResponseError{"api.NextDate Error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
			err = serverJob.UpdateDateTask(idDone, newDateString)
			if err != nil {
				fmt.Println("не удалось обновить дату", err)
				msg, errInt := jsonErrorMarshal(TaskResponseError{"data update error"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}
		var answear = map[string]any{}
		msg, err := json.Marshal(answear)
		if err != nil {
			msg, errInt := jsonErrorMarshal(TaskResponseError{"error when generating the serialization of the response"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(msg)
	}
}

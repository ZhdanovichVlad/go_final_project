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

// структура для получения задач с фронтэнда и отправки обратно.
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// Структура для возврата в формате JSON ответа при добавлении задачи
type TaskResponseID struct {
	ID string `json:"id"`
}

// описываем интерфес нашей базы данных, что бы мы использовать методы для работы с базой данных
type ServerJob interface {
	AddTask(date string, title string, comment string, repeat string) (string, error)
	GetTasks(NumberOfOuptuTasks int, tasks []Task) ([]Task, error)
	GetTask(id string) (Task, error)
	UpdateTask(task Task) error
	DeleteTask(idTask string) error
	UpdateDateTask(idTask string, newDateString string) error
	SearchTasks(code int, searchQuery string, NumberOfOuptuTasks int) ([]Task, error)
}

// ApiNextDate function for the api.NextDate handle. Sends the user the date when the next task should be performed.
// ApiNextDate функция для ручки api.NextDate. Отправляет пользователю дату когда надо выполнить следующую задачу.
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
// PostTask. Функция-хендлер, добавляющая задания в базу данных
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
		var date string // - создаем дату, которая будет записана в базу данных
		// делаем проверку на значение даты и правила повтора.
		// В зависимости от установленных правил устанавливаем date которе будет записано в базу данных
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
		case task.Repeat == "" && task.Date != "":
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

		//// добавляем в БД новую задачу
		lastID, err := serverJob.AddTask(date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		var answear TaskResponseID = TaskResponseID{lastID}

		answearJSON, err := json.Marshal(&answear)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"dicerelization error"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(answearJSON)

	}
}

// GetTasksHundler function to get the nearest tasks. The number of tasks is set by the variable NumberOfOupTasksString. The variable is defined in the .env file
// this function also handles search by date or text.
// GetTasksHundler функция для нахождения ближайших задач. Количество задач задается переменной NumberOfOutTasksString. Переменная определена в файле .env
// в данной функции так же обрабатывается поиск по дате или тексту.
func GetTasksHundler(serverJob ServerJob, NumberOfOutTasksString string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		NumberOfOutTasksInt, err := strconv.Atoi(NumberOfOutTasksString)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"NumberOfOutTasksString to NumberOfOutTasksInt convert error"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		reqQuery := req.URL.Query().Get("search")
		// Исходя из значений reqQuery у нас есть три возможных случая. Каждый будет иметь свою кодировку
		//1-case. reqQuery - пуста.         Код 0. Возвращается заданное в параметре NumberOfOutTasksString ближайших задач
		//2-case. reqQuery содержит дату  - Код 1. Происходит поиск в базе данных по дате.
		//3-case. reqQuery содержит текст - Код 2. Происходит поиск в базе данных по полям title и comment
		var code int
		_, err = time.Parse("02.01.2006", reqQuery)
		if reqQuery == "" {
			code = 0
		} else if err == nil {
			code = 1
		} else {
			code = 2
		}

		var answearRows []Task // создание будущего ответа пользователю
		if code == 0 {         // случай при возврате ближайших задач
			answearRows, err = serverJob.GetTasks(NumberOfOutTasksInt, answearRows)
			if err != nil {
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when retrieving tasks from the database"}, false)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		} else { // случай поиска по дате или по тексту
			answearRows, err = serverJob.SearchTasks(code, reqQuery, NumberOfOutTasksInt)
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

// GetTaskHundler function returns a task from the database by the specified ID in the Query
// GetTaskHundler функция возвращает задачу из базы данных по заданному ID в Query
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
		answerRow, err := serverJob.GetTask(qeryID)
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"error when retrieving tasks from the database"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		jsonMsg, err := json.Marshal(answerRow)
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

// CorrectTask  function that CorrectTask in the database
// CorrectTask функция для корректировки задачи в базе данных по полученному ID и
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

		_, err = strconv.Atoi(task.ID) // проверка, что в поле ID передана цифра
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

		if task.Title == "" { // значение Title всегда должно содержать описание задачи. Если оно пустое возвращаем ошибку

			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"The title is empty"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		} // дата задания должна больше чем сегодня. В противном случае возвращаем ошибку
		if dateTime.Before(time.Now()) && dateTime.Equal(time.Now()) {
			//if dateTime.Before(time.Now()) || !dateTime.Equal(time.Now()) {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"the date can't be less than today"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		// Правило повторение(Repeat) может быть пустой строкой или начинаться из следующих символов y,d,m,w.
		// Если в Repeat указано другое мы возвращаем ошибку.
		// Сразу проверяем, что строка пуста. Если нет, делаем
		// Преобразуем строку в массив рун и сравним первую руну.

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
				msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"the rule for repetition has the wrong format."}, true)
				w.WriteHeader(errInt)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(msg)
				return
			}
		}

		// обновляем задачу в базе данных
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

// DoneTaskHundler function to delete and update a task in the database by a given ID if it has been executed
// DoneTaskHundler функция для удаления и обновления задачи в базе данных по заданному ID, если оно было выполнено
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
		_, err := strconv.Atoi(idDone) // проверка, на то что в поле ID передано число
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}
		task, err := serverJob.GetTask(idDone) // проверка, что задание существует в базе данных
		if task.ID == "" {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"id not found in the database"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return

		}
		// Eсли значение для повторения не указано в базе данных. Мы его удаляем.
		// В противном случае (блок else), рассчитываем новую дату и обновляем задачу
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

// DeleteTaskHundler function to delete a task by the given Query id.
// DeleteTaskHundler функция для удаления задачи по заданному в Query id
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
		_, err := strconv.Atoi(idDone) // проверка, передано ли в поле ID цифра
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"wrong ID"}, true)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		err = serverJob.DeleteTask(idDone) // Удаляем задачу. Если удалить не удалось, возращается ошибка
		if err != nil {
			msg, errInt := http_server.JsonErrorMarshal(http_server.TaskResponseError{"failed to delete the task"}, false)
			w.WriteHeader(errInt)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(msg)
			return
		}

		// формируем и фозвращаем пустой json в качестве ответа
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

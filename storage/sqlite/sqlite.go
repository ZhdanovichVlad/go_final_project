package sqlite

// пакет для работы с базой данных
import (
	"fmt"
	"github.com/ZhdanovichVlad/go_final_project/http-server/handlers"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"strconv"
	"time"
)

type Storage struct {
	db *sqlx.DB
}

// New Функция возвращает ссылку на базу данных. Необходимо передать путь к БД через переменную DBFile
// Если база данных существует, то подключается к существующей базе денных.
// Если база данных не существует, то создает новую БД.
// New The function returns a reference to the database. It is necessary to pass the path to the database through the DBFile variable
// If the database exists, it connects to the existing database.
// If the database does not exist, it creates a new database.
func New(DBFile string) (*Storage, error) {

	_, err := os.Stat(DBFile)
	var install bool
	if err != nil {
		install = true
	}

	if install {
		_, err = os.Create(DBFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	//db, err := sqlx.Connect("sqlite3", DBFile) -- для работы необходим  TDM-GCC
	db, err := sqlx.Connect("sqlite3", DBFile)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if install {
		_, err = db.Exec("CREATE TABLE scheduler (id  INTEGER PRIMARY KEY AUTOINCREMENT, date VARCHAR, title VARCHAR(128) NOT NULL, comment VARCHAR, repeat VARCHAR(128) )")
		if err != nil {
			log.Panic(err)
			return nil, err
		}
		_, err = db.Exec("CREATE INDEX ID_Date ON scheduler (date)")
		if err != nil {
			log.Panic(err)
			return nil, err
		}

	}

	return &Storage{db: db}, nil
}

// Close method to close the database
// Close метод для закрытия базы данных
func (s *Storage) Close() {
	s.db.Close()
}

// AddTask method adds tasks to the database. The method interacts with http-servet.handlers.PostTask .
// Takes data as input and returns id in string format or error
// AddTask метод добавляет задачи в базу данных. Метод взаимодействует с ручкой http-servet.handlers.PostTask .
// Принимает на вход данные и возвращает id в формате string или ошибку
func (s *Storage) AddTask(date string, title string, comment string, repeat string) (string, error) {
	stmt, err := s.db.Prepare("INSERT INTO scheduler(date, title,comment,repeat) VALUES(?, ?,?,?)")
	if err != nil {
		return "", fmt.Errorf("failed to create a request for database update", err)
	}

	res, err := stmt.Exec(date, title, comment, repeat)
	if err != nil {
		return "", fmt.Errorf("failed to INSERT a request for database update", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("Failed to get last insert id:", err)
	}
	idString := strconv.Itoa(int(id))
	return idString, nil

}

// GetTasks method returns the specified in NumberOfOutTasks from the database The method interacts with the handle http-servet.handlers.GetTasksHundler
// Returns a response in the form of a slice of handlers.Task structures or an error
// GetTasks метод возвращает из базы данных указанное в NumberOfOutTasks Метод взаимодействует с ручкой http-servet.handlers.GetTasksHundler
// Возвращает ответ в форме слайса структур handlers.Task или ошибку
func (s Storage) GetTasks(NumberOfOutTasks int, tasks []handlers.Task) ([]handlers.Task, error) {

	stmt, err := s.db.Prepare("SELECT * FROM scheduler ORDER BY date LIMIT ? ")
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for select from database", err)
	}
	rows, err := stmt.Query(NumberOfOutTasks)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("failed request for select from database", err)
	}

	for rows.Next() {
		task := handlers.Task{}
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed scan from database", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetTask method returns a task from the database by the given ID.  The method interacts with http-servet.handlers.GetTaskHundler handle
// Returns a response in the form of handlers.Task structure or an error
// GetTask метод возвращает из базы данных задачу по заданному ID.  Метод взаимодействует с ручкой http-servet.handlers.GetTaskHundler
// Возвращает ответ в форме структуры handlers.Task или ошибку
func (s Storage) GetTask(id string) (handlers.Task, error) {
	stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE id =? ")
	if err != nil {
		return handlers.Task{}, fmt.Errorf("failed to create a request for select from database", err)
	}
	rows, err := stmt.Query(id)
	defer rows.Close()
	if err != nil {
		return handlers.Task{}, fmt.Errorf("failed request for select from database", err)
	}

	task := handlers.Task{}
	for rows.Next() {
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return handlers.Task{}, fmt.Errorf("failed scan from database", err)
		}
	}
	if task.ID == "" {
		return handlers.Task{}, fmt.Errorf("database query not found", err)
	}
	return task, nil
}

// UpdateTask method updates tasks in the database. The method updates all fields. Takes a task as input and returns an error if it failed to update.
// The method interacts with the http-servet.handlers.CorrectTask handle
// UpdateTask метод обновляет в базе данных задачи. Метод обновляет все поля. Принимает на вход задачу и возвращает ошибку, если не удалось обновить.
// Метод взаимодействует с ручкой http-servet.handlers.CorrectTask
func (s Storage) UpdateTask(task handlers.Task) error {
	stmt, err := s.db.Prepare("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?  WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for select from database", err)
	}
	result, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("data update error", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0", err)
	}
	return nil
}

// DeleteTask method for deleting a completed task from the database. Takes task ID as input and returns an error if deletion failed.
// The method interacts with the http-servet.handlers.DeleteTaskHundler handle
// DeleteTask метод для удаления из базы данных выполненной задачи. Принимает на вход ID задачи и возвращает ошибку, если удаление не удалось.
// Метод взаимодействует с ручкой http-servet.handlers.DeleteTaskHundler
func (s Storage) DeleteTask(idTask string) error {
	stmt, err := s.db.Prepare("DELETE FROM scheduler WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for delete from database", err)
	}
	result, err := stmt.Exec(idTask)
	if err != nil {
		return fmt.Errorf("data delete error", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0", err)
	}
	return nil
}

// UpdateDateTask method updates the task database. The method updates only the date. It is needed to correctly change the task if it is repeated.
// Accepts ID and new date as input and returns an error if the update failed.
// The method interacts with the http-servet.handlers.DoneTaskHundler handle
// UpdateDateTask метод обновляет в базе данных задачи. Метод обновляет только дату. Нужен для корректного изменения задачи, если она повторяется.
// Принимает на вход ID и новую дату и возвращает ошибку, если не удалось обновить.
// Метод взаимодействует с ручкой http-servet.handlers.DoneTaskHundler
func (s Storage) UpdateDateTask(idTask string, newDateString string) error {
	stmt, err := s.db.Prepare("UPDATE scheduler SET date = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for update date from database", err)
	}
	result, err := stmt.Exec(newDateString, idTask)
	if err != nil {
		return fmt.Errorf("date update error", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0", err)
	}
	return nil
}

// SearchTasks method searches the database for information by ID or title and comment fields. It accepts encoding 1 or 2 as input,
// which specifies how to search. The method interacts with http-servet.handlers.GetTasksHundler. The encodings are assigned in the same method.
// The method takes as input the code, the string to search for, and the number of handlers to output. Returns a slice of handlers.Task structures or an error.
func (s Storage) SearchTasks(code int, searchQuery string, NumberOfOutTasks int) ([]handlers.Task, error) {
	var tasks []handlers.Task
	switch code {
	case 1:
		date, err := time.Parse("02.01.2006", searchQuery)
		if err != nil {
			return nil, fmt.Errorf("error in date conversion in the searsh function. package sqlite")
		}

		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE date = ? LIMIT ?")

		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database", err)
		}
		rows, err := stmt.Query(date.Format("20060102"), NumberOfOutTasks)
		if err != nil {
			return nil, fmt.Errorf("failed request for select from database", err)
		}
		for rows.Next() {
			task := handlers.Task{}
			err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				return nil, fmt.Errorf("failed scan from database", err)
			}
			tasks = append(tasks, task)
		}
		defer rows.Close()

	case 2:
		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?")
		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database", err)
		}
		rows, err := stmt.Query("%"+searchQuery+"%", "%"+searchQuery+"%", NumberOfOutTasks)
		if err != nil {
			return nil, fmt.Errorf("failed request for select from database", err)
		}
		for rows.Next() {
			task := handlers.Task{}
			err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				return nil, fmt.Errorf("failed scan from database", err)
			}
			tasks = append(tasks, task)
		}

	}
	return tasks, nil
}

package sqlite

// Database package
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

	db, err := sqlx.Connect("sqlite3", DBFile) //-- TDM-GCC is required for operation

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
func (s *Storage) Close() {
	s.db.Close()
}

// AddTask method adds tasks to the database. The method interacts with http-servet.handlers.PostTask .
// Takes data as input and returns id in string format or error
func (s *Storage) AddTask(date string, title string, comment string, repeat string) (string, error) {
	stmt, err := s.db.Prepare("INSERT INTO scheduler(date, title,comment,repeat) VALUES(?, ?,?,?)")
	if err != nil {
		return "", fmt.Errorf("failed to create a request for database update: %w", err)
	}

	res, err := stmt.Exec(date, title, comment, repeat)
	if err != nil {
		return "", fmt.Errorf("failed to INSERT a request for database update: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("Failed to get last insert id: %w", err)
	}
	idString := strconv.Itoa(int(id))
	return idString, nil

}

// GetTasks method returns the specified in NumberOfOutPutTasks from the database The method interacts with the handle http-servet.handlers.GetTasksHundler
// Returns a response in the form of a slice of handlers.Task structures or an error
func (s Storage) GetTasks(NumberOfOutPutTasks int) ([]handlers.Task, error) {
	var tasks []handlers.Task
	stmt, err := s.db.Prepare("SELECT * FROM scheduler ORDER BY date LIMIT ? ")
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for select from database: %w", err)
	}
	rows, err := stmt.Query(NumberOfOutPutTasks)
	if err != nil {
		return nil, fmt.Errorf("failed request for select from database: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		task := handlers.Task{}
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed scan from database: %w", err)
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Next() contains errors: %w", err)
	}
	return tasks, nil
}

// GetTask method returns a task from the database by the given ID.  The method interacts with http-servet.handlers.GetTaskHundler handle
// Returns a response in the form of handlers.Task structure or an error
func (s Storage) GetTask(id string) (handlers.Task, error) {
	stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE id =? ")
	if err != nil {
		return handlers.Task{}, fmt.Errorf("failed to create a request for select from database: %w", err)
	}
	rows, err := stmt.Query(id)
	if err != nil {
		return handlers.Task{}, fmt.Errorf("failed request for select from database: %w", err)
	}
	defer rows.Close()

	task := handlers.Task{}
	for rows.Next() {
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return handlers.Task{}, fmt.Errorf("failed scan from database: %w", err)
		}
	}
	if err != nil {
		return handlers.Task{}, fmt.Errorf("rows.Next() contains errors: %w", err)
	}
	if task.ID == "" {
		return handlers.Task{}, fmt.Errorf("database query not found: %w", err)
	}
	return task, nil
}

// UpdateTask method updates tasks in the database. The method updates all fields. Takes a task as input and returns an error if it failed to update.
// The method interacts with the http-servet.handlers.CorrectTask handle
func (s Storage) UpdateTask(task handlers.Task) error {
	stmt, err := s.db.Prepare("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?  WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for select from database: %w", err)
	}
	result, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("data update error: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0: %w", err)
	}
	return nil
}

// DeleteTask method for deleting a completed task from the database. Takes task ID as input and returns an error if deletion failed.
// The method interacts with the http-servet.handlers.DeleteTaskHundler handle
func (s Storage) DeleteTask(idTask string) error {
	stmt, err := s.db.Prepare("DELETE FROM scheduler WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for delete from database: %w", err)
	}
	result, err := stmt.Exec(idTask)
	if err != nil {
		return fmt.Errorf("data delete error: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0: %w", err)
	}
	return nil
}

// UpdateDateTask method updates the task database. The method updates only the date. It is needed to correctly change the task if it is repeated.
// Accepts ID and new date as input and returns an error if the update failed.
// The method interacts with the http-servet.handlers.DoneTaskHundler handle
func (s Storage) UpdateDateTask(idTask string, newDateString string) error {
	stmt, err := s.db.Prepare("UPDATE scheduler SET date = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to create a request for update date from database: %w", err)
	}
	result, err := stmt.Exec(newDateString, idTask)
	if err != nil {
		return fmt.Errorf("date update error: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error when receiving information about the number of updated rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("the number of updated tasks is 0: %w", err)
	}
	return nil
}

// SearchTasks method searches the database for information by ID or title and comment fields. It accepts encoding 1 (handlers.DateSearch) or 2(handlers.TextSearch) as input,
// which specifies how to search. The method interacts with http-servet.handlers.GetTasksHundler. The encodings are assigned in the same method.
// The method takes as input the code, the string to search for, and the number of handlers to output. Returns a slice of handlers.Task structures or an error.
func (s Storage) SearchTasks(code int, searchQuery string, NumberOfOutPutTasks int) ([]handlers.Task, error) {
	var tasks []handlers.Task
	switch code {
	case handlers.DateSearch:
		date, err := time.Parse("02.01.2006", searchQuery)
		if err != nil {
			return nil, fmt.Errorf("error in date conversion in the searsh function. package sqlite: %w", err)
		}

		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE date = ? LIMIT ?")

		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database: %w", err)
		}
		rows, err := stmt.Query(date.Format("20060102"), NumberOfOutPutTasks)
		if err != nil {
			return nil, fmt.Errorf("failed request for select from database: %w", err)
		}
		for rows.Next() {
			task := handlers.Task{}
			err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				return nil, fmt.Errorf("failed scan from database: %w", err)
			}
			tasks = append(tasks, task)
		}
		if err != nil {
			return nil, fmt.Errorf("rows.Next() contains errors: %w", err)
		}
		defer rows.Close()

	case handlers.TextSearch:
		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?")
		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database: %w", err)
		}
		rows, err := stmt.Query("%"+searchQuery+"%", "%"+searchQuery+"%", NumberOfOutPutTasks)
		if err != nil {
			return nil, fmt.Errorf("failed request for select from database: %w", err)
		}
		for rows.Next() {
			task := handlers.Task{}
			err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				return nil, fmt.Errorf("failed scan from database: %w", err)
			}
			tasks = append(tasks, task)
		}
		if err != nil {
			return nil, fmt.Errorf("rows.Next() contains errors: %w", err)
		}

	}
	return tasks, nil
}

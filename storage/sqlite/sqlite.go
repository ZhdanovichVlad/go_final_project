package sqlite

import (
	"fmt"
	"github.com/ZhdanovichVlad/go_final_project/http-server/handlers"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"time"
)

type Storage struct {
	db *sqlx.DB
}

// New Функция возвращает ссылку на базу данных. Необходимо передать путь к БД через переменную DBFile
// Если база данных существует, то подключается к существующей базе денных.
// Если база данных не существует, то создает новую БД.
func New(DBFile string) (*Storage, error) {
	//appPath, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//DBFile := filepath.Join(appPath, "scheduler.db")
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

	//db, err := sql.Open("sqlite3", DBFile)
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

func (s *Storage) AddTask(date string, title string, comment string, repeat string) (int64, error) {
	stmt, err := s.db.Prepare("INSERT INTO scheduler(date, title,comment,repeat) VALUES(?, ?,?,?)")
	if err != nil {
		return 0, fmt.Errorf("failed to create a request for database update", err)
	}

	res, err := stmt.Exec(date, title, comment, repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to INSERT a request for database update", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Failed to get last insert id:", err)
	}
	return id, nil

}

func (s *Storage) Close() {
	s.db.Close()
}

func (s Storage) GetTasks(NumberOfOuptuTasks int, tasks []handlers.Task) ([]handlers.Task, error) {

	stmt, err := s.db.Prepare("SELECT * FROM scheduler ORDER BY date LIMIT ? ")
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for select from database", err)
	}
	rows, err := stmt.Query(NumberOfOuptuTasks)
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

// GetTask function to retrieve one task. receives task id as input and returns Task structure and error.
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

func (s Storage) SearchTasks(code int, searchQuery string, NumberOfOuptuTasks int) ([]handlers.Task, error) {
	fmt.Println("вход в SearchTasks")
	fmt.Println("Значение code ", code)
	var tasks []handlers.Task
	switch code {
	case 1:
		fmt.Println("Вход в case 1")
		date, err := time.Parse("02.01.2006", searchQuery)
		if err != nil {
			return nil, fmt.Errorf("error in date conversion in the searsh function. package sqlite")
		}

		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE date = ? LIMIT ?")

		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database", err)
		}
		rows, err := stmt.Query(date.Format("20060102"), NumberOfOuptuTasks)
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
		fmt.Println("Вход в case 2")
		stmt, err := s.db.Prepare("SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?")
		if err != nil {
			return nil, fmt.Errorf("failed to create a request for select from database", err)
		}
		rows, err := stmt.Query("%"+searchQuery+"%", "%"+searchQuery+"%", NumberOfOuptuTasks)
		if err != nil {
			return nil, fmt.Errorf("failed request for select from database", err)
		}
		for rows.Next() {
			task := handlers.Task{}
			err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			fmt.Println("найдена строка ", task)
			if err != nil {
				return nil, fmt.Errorf("failed scan from database", err)
			}
			tasks = append(tasks, task)
		}
		fmt.Println("tasks в поиске", tasks)

	}
	return tasks, nil
}

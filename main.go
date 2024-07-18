package main

import (
	"fmt"
	"github.com/ZhdanovichVlad/go_final_project/http-server/handlers"
	"github.com/ZhdanovichVlad/go_final_project/storage/sqlite"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	var port string = os.Getenv("TODO_PORT")
	var DBFile = os.Getenv("TODO_DBFILE")
	var NumberOfOuptuTasks = os.Getenv("NUMNER_OF_OUTPUT_TASKS")

	storage, err := sqlite.New(DBFile)
	//_ = storage
	defer storage.Close()

	server := chi.NewRouter()

	server.Handle("/", http.FileServer(http.Dir("web")))
	server.HandleFunc("/api/nextdate", handlers.ApiNextDate)
	server.Post("/api/task", handlers.PostTask(storage))
	server.Post("/api/task/done", handlers.DoneTaskHundler(storage))
	server.Get("/api/task", handlers.GetTaskHundler(storage))
	server.Put("/api/task", handlers.CorrectTask(storage))
	server.Get("/api/tasks", handlers.GetTasksHundler(storage, NumberOfOuptuTasks))

	//server.Post("/api/task", handlers.PostTask)

	fmt.Println("Сервер Запускается ")

	err = http.ListenAndServe("localhost:"+port, server)
	if err != nil {
		log.Panic(err)
	}

}

//func handlerApiNextDate(w http.ResponseWriter, req *http.Request) {
//	now := req.FormValue("now")
//	nowTime, err := time.Parse("20060102", now)
//	if err != nil {
//		log.Panic(err)
//	}
//	date := req.FormValue("date")
//	repeat := req.FormValue("repeat")
//	answer, err := api.NextDate(nowTime, date, repeat)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	w.Write([]byte(answer))
//
//}

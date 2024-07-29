package main

import (
	authorization "github.com/ZhdanovichVlad/go_final_project/http-server/auth"
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
	// используем библиотеку godotenv что бы загрузить переменные окружения корня проекта
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	var port string = os.Getenv("TODO_PORT")
	var DBFile = os.Getenv("TODO_DBFILE")
	var NumberOfOuptuTasks = os.Getenv("NUMNER_OF_OUTPUT_TASKS")

	storage, err := sqlite.New(DBFile)
	defer storage.Close()

	server := chi.NewRouter()
	// Описание основных хэндлеров. Middleware authorization.CheckToken используется для проверки JWT токена
	// Токен формируется при авторизации пользователя.
	server.Handle("/*", http.FileServer(http.Dir("web")))
	server.HandleFunc("/api/nextdate", handlers.ApiNextDate)
	server.Post("/api/task", authorization.CheckToken(handlers.PostTask(storage)))
	server.Get("/api/task", authorization.CheckToken(handlers.GetTaskHundler(storage)))
	server.Put("/api/task", authorization.CheckToken(handlers.CorrectTask(storage)))
	server.Post("/api/task/done", authorization.CheckToken(handlers.DoneTaskHundler(storage)))
	server.Delete("/api/task", authorization.CheckToken(handlers.DeleteTaskHundler(storage)))
	server.Get("/api/tasks", authorization.CheckToken(handlers.GetTasksHundler(storage, NumberOfOuptuTasks)))
	server.Post("/api/signin", authorization.Authorization)

	log.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		log.Fatal(err)
	}

}

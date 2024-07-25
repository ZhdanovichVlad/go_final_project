package main

import (
	"fmt"
	authorization "github.com/ZhdanovichVlad/go_final_project/http-server/auth"
	"github.com/ZhdanovichVlad/go_final_project/http-server/handlers"
	"github.com/ZhdanovichVlad/go_final_project/http-server/midll"
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
	defer storage.Close()

	service := authorization.NewService()
	authMiddl := midll.NewAuthMiddleware(service)

	server := chi.NewRouter()
	server.Handle("/*", http.FileServer(http.Dir("web")))
	//server.HandleFunc("/api/nextdate", handlers.ApiNextDate)
	//server.Post("/api/task", handlers.PostTask(storage))
	//server.Get("/api/task", handlers.GetTaskHundler(storage))
	//server.Put("/api/task", handlers.CorrectTask(storage))
	//server.Post("/api/task/done", handlers.DoneTaskHundler(storage))
	//server.Delete("/api/task", handlers.DeleteTaskHundler(storage))
	//server.Get("/api/tasks", handlers.GetTasksHundler(storage, NumberOfOuptuTasks))
	server.Get("/api/tasks", authMiddl.CheckToken(handlers.GetTasksHundler(storage, NumberOfOuptuTasks)))
	server.Post("/api/signin", service.Authorization)

	fmt.Println("запускается сервер")
	err = http.ListenAndServe("localhost:"+port, server)
	if err != nil {
		log.Panic(err)
	}

}

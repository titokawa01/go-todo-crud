package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
)

var db *sql.DB

type PageData struct {
	Message string
	Tasks []Task
}

type Task struct {
	ID int `json:"id"`
	Content string `json:"content"`
}

func main() {

	var err error

	db, err = sql.Open("mysql", "root:9999@unix(/tmp/mysql.sock)/mermaid")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("DB接続成功")
	http.HandleFunc("/", handler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/api/tasks", apiTasksHandler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    var name string

	if r.Method == "POST"{
		name := r.FormValue("name")

		_, err := db.Exec("INSERT INTO tasks (content) VALUES (?)", name)
		if err != nil {
			panic(err)
		}
	}

	rows, err := db.Query ("SELECT id, content FROM tasks")
	if err != nil {
		panic(err)
	}
    defer rows.Close()

	var tasks []Task

    for rows.Next(){
	var t Task
	err := rows.Scan(&t.ID, &t.Content)
	if err != nil {
		panic(err)
	}
		tasks = append(tasks, t)
}	

		data := PageData{
			Message: name,
			Tasks: tasks,
		}
	
		tmpl.Execute(w, data)
	}

func deleteHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")

	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r * http.Request){
	id := r.URL.Query().Get("id")

	row := db.QueryRow("SELECT id, content FROM tasks WHERE id = ?", id)

	var task Task
	err := row.Scan(&task.ID, &task.Content)
	if err != nil {
		panic(err)
	}
	tmpl, _ := template.ParseFiles("templates/edit.html")
	tmpl.Execute(w, task)
}

func updateHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	content := r.FormValue("content")

	_, err := db.Exec("UPDATE tasks SET content = ? WHERE id = ?", content, id)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func apiTasksHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println("Method:", r.Method)
	if r.Method == "POST" {
		var t Task

		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return	
	}

	result, err := db.Exec("INSERT INTO tasks (content) VALUES (?)", t.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)
    
    w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
	return
    }

	if r.Method == "PUT"{
		fmt.Println("PUTに入りました")
		id := r.URL.Query().Get("id")

		var t Task
		err := json.NewDecoder(r.Body).Decode(&t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE tasks SET content = ? WHERE id = ?", t.Content, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == "DELETE"{
		id := r.URL.Query().Get("id")

		_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	rows, err := db.Query("SELECT id, content FROM tasks")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next(){
		var t Task
		if err := rows.Scan(&t.ID, &t.Content); err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
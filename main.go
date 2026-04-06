package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strconv"

    "github.com/gorilla/sessions"
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
)
// Структури
type Generator struct {
	ID     int
	Name   string
	Power  int
	Status string
}

type Consumer struct {
	ID     int
	Name   string
	Load   int
	Status string
}

type Sensor struct {
	ID    int
	Type  string
	Value int
	Unit  string
}

var db *sql.DB
var store = sessions.NewCookieStore([]byte("secret-key"))

func main() {
	var err error
	dsn := "poweruser:powerpass@tcp(127.0.0.1:3306)/powerdb"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Помилка підключення:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("База недоступна:", err)
	}
	fmt.Println("Підключення успішне!")

	// Маршрути
	http.HandleFunc("/", homePage)

	http.HandleFunc("/generators", listGenerators)
	http.HandleFunc("/generators/add", addGenerator)
	http.HandleFunc("/generators/edit", editGenerator)
	http.HandleFunc("/generators/delete", deleteGenerator)

	http.HandleFunc("/consumers", listConsumers)
	http.HandleFunc("/consumers/add", addConsumer)
	http.HandleFunc("/consumers/edit", editConsumer)
	http.HandleFunc("/consumers/delete", deleteConsumer)

	http.HandleFunc("/sensors", listSensors)
	http.HandleFunc("/sensors/add", addSensor)
	http.HandleFunc("/sensors/edit", editSensor)
	http.HandleFunc("/sensors/delete", deleteSensor)

	// Auth
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/profile", profile)
	http.HandleFunc("/logout", logout)

	// Swagger UI (роздача статичних файлів з папки swagger)
	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))

	log.Println("Сервер запущено на :8080")
	http.ListenAndServe(":8080", nil)
}

// ---------------- HTML шаблони ----------------
var tmplHome = template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="uk">
<head><meta charset="UTF-8"><title>Smart Grid</title></head>
<body>
    <h1>Smart Grid CRUD</h1>
    <ul>
        <li><a href="/generators">Generators</a></li>
        <li><a href="/consumers">Consumers</a></li>
        <li><a href="/sensors">Sensors</a></li>
    </ul>

    <h2>User Section</h2>
    <ul>
        <li><a href="/register">Register</a></li>
        <li><a href="/login">Login</a></li>
        <li><a href="/profile">Profile</a></li>
        <li><a href="/logout">Logout</a></li>
    </ul>

    <h2>API Documentation</h2>
    <ul>
        <li><a href="/swagger/index.html">Swagger UI</a></li>
    </ul>
</body>
</html>
`))

var tmplList = template.Must(template.New("list").Parse(`
<!DOCTYPE html>
<html lang="uk">
<head><meta charset="UTF-8"><title>Smart Grid</title></head>
<body>
    <h1>{{.Title}}</h1>
    <table border="1" cellpadding="5">
        <tr>
            {{range .Headers}}<th>{{.}}</th>{{end}}
            <th>Actions</th>
        </tr>
        {{range .Rows}}
        <tr>
            {{range .}}<td>{{.}}</td>{{end}}
            <td>
                <a href="{{$.EditURL}}?id={{index . 0}}">Edit</a> |
                <a href="{{$.DeleteURL}}?id={{index . 0}}">Delete</a>
            </td>
        </tr>
        {{end}}
    </table>

    <h2>Add {{.Title}}</h2>
    <form method="POST" action="{{.AddURL}}">
        {{range .FormFields}}{{.}}: <input type="text" name="{{.}}"><br>{{end}}
        <input type="submit" value="Add">
    </form>
    <p><a href="/">Back</a></p>
</body>
</html>
`))

// ---------------- Handlers ----------------
func homePage(w http.ResponseWriter, r *http.Request) {
	tmplHome.Execute(w, nil)
}

// Generators
func listGenerators(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, power, status FROM generators")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var data [][]interface{}
	for rows.Next() {
		var g Generator
		if err := rows.Scan(&g.ID, &g.Name, &g.Power, &g.Status); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data = append(data, []interface{}{g.ID, g.Name, g.Power, g.Status})
	}
	tmplList.Execute(w, map[string]interface{}{
		"Title":      "Generators",
		"Headers":    []string{"ID", "Name", "Power", "Status"},
		"Rows":       data,
		"AddURL":     "/generators/add",
		"EditURL":    "/generators/edit",
		"DeleteURL":  "/generators/delete",
		"FormFields": []string{"name", "power", "status"},
	})
}

func addGenerator(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		power, _ := strconv.Atoi(r.FormValue("power"))
		status := r.FormValue("status")
		db.Exec("INSERT INTO generators (name, power, status) VALUES (?, ?, ?)", name, power, status)
		http.Redirect(w, r, "/generators", http.StatusSeeOther)
	}
}

func editGenerator(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if r.Method == http.MethodPost {
		status := r.FormValue("status")
		db.Exec("UPDATE generators SET status=? WHERE id=?", status, id)
		http.Redirect(w, r, "/generators", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<form method="POST" action="/generators/edit?id=%d">
        New Status: <input type="text" name="status"><br>
        <input type="submit" value="Update">
    </form>`, id)
}

func deleteGenerator(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	db.Exec("DELETE FROM generators WHERE id=?", id)
	http.Redirect(w, r, "/generators", http.StatusSeeOther)
}

// Consumers
func listConsumers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, `load`, status FROM consumers")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var data [][]interface{}
	for rows.Next() {
		var c Consumer
		if err := rows.Scan(&c.ID, &c.Name, &c.Load, &c.Status); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data = append(data, []interface{}{c.ID, c.Name, c.Load, c.Status})
	}
	tmplList.Execute(w, map[string]interface{}{
		"Title":      "Consumers",
		"Headers":    []string{"ID", "Name", "Load", "Status"},
		"Rows":       data,
		"AddURL":     "/consumers/add",
		"EditURL":    "/consumers/edit",
		"DeleteURL":  "/consumers/delete",
		"FormFields": []string{"name", "load", "status"},
	})
}

func addConsumer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		load, _ := strconv.Atoi(r.FormValue("load"))
		status := r.FormValue("status")
		db.Exec("INSERT INTO consumers (name, `load`, status) VALUES (?, ?, ?)", name, load, status)
		http.Redirect(w, r, "/consumers", http.StatusSeeOther)
	}
}

func editConsumer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if r.Method == http.MethodPost {
		status := r.FormValue("status")
		db.Exec("UPDATE consumers SET status=? WHERE id=?", status, id)
		http.Redirect(w, r, "/consumers", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<form method="POST" action="/consumers/edit?id=%d">
        New Status: <input type="text" name="status"><br>
        <input type="submit" value="Update">
    </form>`, id)
}

func deleteConsumer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	db.Exec("DELETE FROM consumers WHERE id=?", id)
	http.Redirect(w, r, "/consumers", http.StatusSeeOther)
}

// Sensors
func listSensors(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, type, value, unit FROM sensors")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var data [][]interface{}
	for rows.Next() {
		var s Sensor
		if err := rows.Scan(&s.ID, &s.Type, &s.Value, &s.Unit); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data = append(data, []interface{}{s.ID, s.Type, s.Value, s.Unit})
	}
	tmplList.Execute(w, map[string]interface{}{
		"Title":      "Sensors",
		"Headers":    []string{"ID", "Type", "Value", "Unit"},
		"Rows":       data,
		"AddURL":     "/sensors/add",
		"EditURL":    "/sensors/edit",
		"DeleteURL":  "/sensors/delete",
		"FormFields": []string{"type", "value", "unit"},
	})
}

func addSensor(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sType := r.FormValue("type")
		value, _ := strconv.Atoi(r.FormValue("value"))
		unit := r.FormValue("unit")
		db.Exec("INSERT INTO sensors (type, value, unit) VALUES (?, ?, ?)", sType, value, unit)
		http.Redirect(w, r, "/sensors", http.StatusSeeOther)
	}
}

func editSensor(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if r.Method == http.MethodPost {
		value, _ := strconv.Atoi(r.FormValue("value"))
		db.Exec("UPDATE sensors SET value=? WHERE id=?", value, id)
		http.Redirect(w, r, "/sensors", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<form method="POST" action="/sensors/edit?id=%d">
        New Value: <input type="number" name="value"><br>
        <input type="submit" value="Update">
    </form>`, id)
}

func deleteSensor(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	db.Exec("DELETE FROM sensors WHERE id=?", id)
	http.Redirect(w, r, "/sensors", http.StatusSeeOther)
}

// ---------------- Auth ----------------
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hash)
		if err != nil {
			http.Error(w, "User exists", 400)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<form method="POST" action="/register">
        Username: <input type="text" name="username"><br>
        Password: <input type="password" name="password"><br>
        <input type="submit" value="Register">
    </form>`)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		var hash string
		err := db.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&hash)
		if err != nil {
			http.Error(w, "Invalid user", 401)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			http.Error(w, "Wrong password", 401)
			return
		}
		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<form method="POST" action="/login">
        Username: <input type="text" name="username"><br>
        Password: <input type="password" name="password"><br>
        <input type="submit" value="Login">
    </form>`)
}

func profile(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>User Profile</h1><p>Logged in as: %s</p><p><a href='/'>Back</a></p>", username)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

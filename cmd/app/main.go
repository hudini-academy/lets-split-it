package main

import (
	"expense/pkg/models/mysql"
	"flag"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

// Application holds the application dependencies.
type Application struct {
	Config   *Config
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Session  *sessions.Session
	User     *mysql.UserModel
	Expense  *mysql.SplitModel
	Split    *mysql.SplitModel
}

// Config holds the configuration settings.
type Config struct {
	Addr      string
	StaticDir string
	Dsn       string
}

func main() {
	// initialize the config struct.
	config := new(Config)
	// Set default configuration.
	flag.StringVar(&config.Addr, "addr", ":4000", "Default port to listen on")
	flag.StringVar(&config.StaticDir, "staticDir", "./ui/static", "Directory to serve static files from")
	flag.StringVar(&config.Dsn, "dsn", "root:root@/expensetracker?parseTime=true", "MysqlServer connection")
	flag.Parse()

	// initialize the loggers.
	infoLog, errorLog, errInitLoggers := LogFiles()
	if errInitLoggers != nil {
		log.Println(errInitLoggers)
	}

	// initialize database connection
	db, errInitDb := openDB(config.Dsn)
	if errInitDb != nil {
		log.Println(errInitDb)
	}
	defer db.Close()

	// Initialize session
	session := sessions.New([]byte("s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge"))
	session.Lifetime = time.Hour * 12
	// initialize the application struct.
	app := &Application{
		Config:   config,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Session:  session,
		User:     &mysql.UserModel{DB: db},
		Expense:  &mysql.SplitModel{DB: db},
		Split:    &mysql.SplitModel{DB: db},
	}
	errorLog.Fatal(http.ListenAndServe(app.Config.Addr, app.routes()))
}

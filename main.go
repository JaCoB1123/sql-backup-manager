package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	_ "github.com/denisenkom/go-mssqldb"
	"net/http"
)

type config struct {
	Host                     string
	Instance                 string
	User                     string
	Password                 string
	UseWindowsAuthentication bool
}

type database struct {
	Name          string
	ID            int
	Filename      string
	LogFilename string
	Compatibility int
	Version       *int
}

func main() {
	f, err := os.Open("./config.json")
	if err != nil {
		fmt.Println("Cannot open Config: ", err.Error())
		return
	}

	var config config
	dec := json.NewDecoder(f)
	err = dec.Decode(&config)
	f.Close()

	if err != nil {
		fmt.Println("Cannot decode Config: ", err.Error())
		return
	}

	query := url.Values{}
	query.Add("app name", "sql-backup-manager")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(config.User, config.Password),
		Host:     config.Host,
		Path:     config.Instance,
		RawQuery: query.Encode(),
	}
	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		fmt.Println("Cannot connect: ", err.Error())
		return
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Cannot connect: ", err.Error())
		return
	}
	defer db.Close()

	http.HandleFunc("/", rootHandler(db))
	http.ListenAndServe(":8082", nil)
}

func rootHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		root, tail := ShiftPath(r.URL.Path)
		if root != "api" {
			http.NotFound(w, r)
			return
		}

		tail, _ = ShiftPath(tail)
		switch tail {
		case "databases"		:
			dbs, err := getDatabases(db)
			if err != nil {
				http.Error(w, "Could not serialize", http.StatusInternalServerError)
			}

			enc := json.NewEncoder(w)
			enc.Encode(dbs)
			break
		default:
			fmt.Fprint(w, "Hit " + tail)
			break
		}
	}
}

func getDatabases(db *sql.DB) ([]database, error){
	rows, err := db.Query(
		"SELECT dbs.name, dbid, datafile.physical_name, logfile.physical_name, cmptlevel, version " +
		"FROM sys.sysdatabases dbs " +
		"LEFT JOIN sys.master_files datafile ON datafile.database_id = dbid AND datafile.type = 0" +
		"LEFT JOIN sys.master_files logfile ON logfile.database_id = dbid AND logfile.type = 1")
	if err != nil {
		return nil, err
	}

	var databases []database
	for rows.Next() {
		var database database
		err := rows.Scan(&database.Name, &database.ID, &database.Filename, &database.LogFilename, &database.Compatibility, &database.Version)
		if err != nil {
			return nil, err
		}

		databases = append(databases, database)
	}

	return databases, nil
}

func getDatabaseByName(databases []database, name string) *database {
	for i := range databases {
		db := databases[i]
		if db.Name == name {
			return &db
		}
	}

	return nil
}
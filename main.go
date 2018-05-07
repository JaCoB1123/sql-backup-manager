package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
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
	Compatibility int
	Version       *int
}

func main() {
	f, err := os.Open("./config.json")
	if err != nil {
		panic(err)
	}

	var config config
	dec := json.NewDecoder(f)
	err = dec.Decode(&config)
	f.Close()

	if err != nil {
		panic(err)
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
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Cannot connect: ", err.Error())
		return
	}
	defer db.Close()

	databases, err := getDatabases(db)
	if err != nil {
		panic(err)
	}
	for _, database := range databases {
		fmt.Println(database.Name + ":")
		fmt.Println("  " + database.Filename)
	}
}

func getDatabases(db *sql.DB) ([]database, error){
	rows, err := db.Query("SELECT name, dbid, filename, cmptlevel, version FROM sys.sysdatabases")
	if err != nil {
		return nil, err
	}

	var databases []database
	for rows.Next() {
		var database database
		err := rows.Scan(&database.Name, &database.ID, &database.Filename, &database.Compatibility, &database.Version)
		if err != nil {
			return nil, err
		}

		databases = append(databases, database)
	}

	return databases, nil
}

func printValue(pval *interface{}) {
	switch v := (*pval).(type) {
	case nil:
		fmt.Print("NULL")
	case bool:
		if v {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
	case []byte:
		fmt.Print(string(v))
	case time.Time:
		fmt.Print(v.Format("2006-01-02 15:04:05.999"))
	default:
		fmt.Print(v)
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"strconv"
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

	databases, err := getDatabases(db)
	if err != nil {
		panic(err)
	}

	for i, database := range databases {
		fmt.Printf("%d: %s\n", i+1, database.Name)
	}

	var selectedDatabase *database
	for selectedDatabase == nil {
		var input string
		fmt.Print("Select Database: ")
		fmt.Scanln(&input)

		i, err := strconv.Atoi(input)
		if err == nil && i > 0 && i <= len(databases) {
			selectedDatabase = &databases[i-1]
		} else {
			selectedDatabase = getDatabaseByName(databases, input)
		}
	}

	fmt.Println("Selected Database " + selectedDatabase.Name)
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

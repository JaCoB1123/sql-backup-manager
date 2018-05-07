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
	fmt.Println(u.String())
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

	rows, err := db.Query("SELECT * FROM sys.sysdatabases")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	if cols == nil {
		return
	}

	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
		if i != 0 {
			fmt.Print("\t")
		}
		fmt.Print(cols[i])
	}
	fmt.Println()
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for i := 0; i < len(vals); i++ {
			if i != 0 {
				fmt.Print("\t")
			}
			printValue(vals[i].(*interface{}))
		}
		fmt.Println()

	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
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

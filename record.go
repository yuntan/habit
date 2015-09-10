package main

import (
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"

	"database/sql"
	"strconv"
	"time"
)

const (
	DB_PATH  = "./habit.db"
	DATE_FMT = "20060102"
)

type HabitRecord struct {
	Id   int  `db:"id"`
	Hour int  `db:"hour"`
	Date int  `db:"date"`
	Done bool `db:"done"`
}

func setupDb() (*gorp.DbMap, error) {
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return nil, err
	}
	dbmap := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.SqliteDialect{},
	}
	dbmap.AddTableWithName(HabitRecord{}, "habit").SetKeys(true, "Id")
	// t.ColMap("Id").Rename("id")
	// t.ColMap("Hour").Rename("hour")
	// t.ColMap("Date").Rename("date")
	// t.ColMap("Done").Rename("done")
	return dbmap, dbmap.CreateTablesIfNotExists()
}

func saveRecord(habit Habit, done bool) error {
	d, _ := strconv.Atoi(time.Now().Format(DATE_FMT))
	return dbmap.Insert(&HabitRecord{
		Hour: habit.Hour,
		Date: d,
		Done: done,
	})
}

func getRecord(year int, month time.Month) []HabitRecord {
	d := year*10000 + int(month)*100
	var rcrds []HabitRecord
	_, err := dbmap.Select(&rcrds,
		"select * from habit where date between ? and ? order by date, hour",
		d, d+100)
	if err != nil {
		return nil
	}
	return rcrds
}

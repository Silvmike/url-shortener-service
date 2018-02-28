package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"time"
	"math/rand"
	"net/url"
)

const SHORT_LENGTH = 10
const CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS urls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  url VARCHAR(4096),
  short VARCHAR(%d)
)
`
const CREATE_URL_INDEX = "CREATE UNIQUE INDEX IF NOT EXISTS url_idx on urls (url)"
const CREATE_SHORT_INDEX = "CREATE UNIQUE INDEX IF NOT EXISTS short_idx on urls (short)"
const GET_URL_BY_SHORT = "SELECT url FROM urls WHERE short = ?"
const GET_SHORT_BY_URL = "SELECT short FROM urls WHERE url = ?"
const INSERT_URL = "INSERT INTO urls(url, short) VALUES(?, ?)"

type Url struct {
	long string
	short string
}

type UrlDatabase struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
}

func (udb *UrlDatabase) openDatabase(filename string) {
	var err error
	udb.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		panic(err)
	}
	udb.stmts = make(map[string]*sql.Stmt)
}

func (udb *UrlDatabase) createSchema() {
	executeOrPanic := func(query string) {
		_, err := udb.db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
	executeOrPanic(fmt.Sprintf(CREATE_TABLE, SHORT_LENGTH))
	executeOrPanic(CREATE_URL_INDEX)
	executeOrPanic(CREATE_SHORT_INDEX)
}

func (udb *UrlDatabase) Init(filename string) {

	udb.openDatabase(filename)
	udb.createSchema()
	udb.prepareStatements()

}
func (udb *UrlDatabase) prepareStatements() {
	prepareOrPanic := func(query string) {
		var err error
		udb.stmts[query], err = udb.db.Prepare(query)
		if err != nil {
			panic(err)
		}
	}
	prepareOrPanic(GET_URL_BY_SHORT)
	prepareOrPanic(GET_SHORT_BY_URL)
	prepareOrPanic(INSERT_URL)
}

func (udb *UrlDatabase) InsertUrl(object Url) (sql.Result, error) {
	return udb.stmts[INSERT_URL].Exec(object.long, object.short)
}

func (udb *UrlDatabase) FindUrlByLong(long string) *Url {
	row := udb.stmts[GET_SHORT_BY_URL].QueryRow(long)
	val := Url{long, ""}
	err := row.Scan(&(val.short))
	if err == nil {
		return &val
	}
	return nil
}

func (udb *UrlDatabase) FindUrlByShort(short string) *Url {
	row := udb.stmts[GET_URL_BY_SHORT].QueryRow(short)
	val := Url{"", short}
	err := row.Scan(&(val.long))
	if err == nil {
		return &val
	}
	return nil
}

func (udb *UrlDatabase) Close() {
	for _, value := range udb.stmts {
		value.Close()
	}
	udb.db.Close()
}

type Shortener struct {
	udb *UrlDatabase
}

func (s *Shortener) Init(filename string) {
	s.udb = &UrlDatabase{}
	s.udb.Init(filename)
}

func (s *Shortener) Close() {
	s.udb.Close()
}

type NonUniqueShortError struct {
	When time.Time
	What string
}

func (e NonUniqueShortError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

type InvalidUrlError struct {
	When time.Time
	What string
}

func (e InvalidUrlError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func validateUrl(long string) error {
	_, err := url.ParseRequestURI(long)
	if err != nil {
		return InvalidUrlError{time.Now(), fmt.Sprintf("Bad url [ %s ]", long) }
	}
	return nil
}

func (s *Shortener) Shorten(long string) (*Url, error) {
	err := validateUrl(long)
	if err != nil {
		return nil, err
	}
	return shorten(long, func(long string, value string) (string, error) {

		u := s.udb.FindUrlByLong(long)
		if u == nil {
			u := &Url{long, value}
			_, err := s.udb.InsertUrl(*u)
			if err != nil {
				return "", NonUniqueShortError{time.Now(), fmt.Sprintf("Ooops! Unable to generate unique short sequence", value)}
			}
			return value, nil
		} else {
			return u.short, nil
		}

	})
}

type UrlNotFoundError struct {
	When time.Time
	What string
}

func (e UrlNotFoundError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func (s *Shortener) Lookup(short string) (*Url, error) {
	url := s.udb.FindUrlByShort(short)
	if url == nil {
		return nil, UrlNotFoundError{time.Now(), "Given url wasn't found"}
	}
	return url, nil
}

const SYMBOLS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func shorten(long string, check func(long string, value string) (string, error)) (*Url, error) {

	var err error
	var short string

	randUid := make([]byte, SHORT_LENGTH)

	for attemps := 3; attemps>0; attemps-- {

		for i := range randUid {
			randUid[i] = SYMBOLS[rand.Intn(len(SYMBOLS))]
		}
		short, err = check(long, string(randUid))
		if err == nil {
			return &Url{ long,  short }, nil
		}

	}

	return nil, err

}
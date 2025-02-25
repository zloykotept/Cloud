package main

import (
	"Cloud/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	db, err := sqlx.Connect("pgx", "user=postgres password="+os.Getenv("PG_PASSWORD")+" dbname=cloud sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// MIGRATIONS LOAD
	usersSQL, err := os.ReadFile("db/migrations/users_up.sql")
	filesSQL, err := os.ReadFile("db/migrations/files_up.sql")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(string(usersSQL))
	_, err = db.Exec(string(filesSQL))
	if err != nil {
		log.Fatal(err)
	}

	res, _ := db.Exec(`SELECT * FROM users WHERE name = 'admin'`)
	rows, _ := res.RowsAffected()
	if rows == 0 {
		passHash, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
		_, err = db.Exec(`INSERT INTO users (name, password, permissions, space) VALUES ($1, $2, $3, $4)`, "admin", passHash, 1, 5000)
		if err != nil {
			log.Fatal(err)
		}
	}
	db.Exec(`SET TIMEZONE = 'Europe/Kiev'`) // TEMPORARY

	cfg := config{
		addr:              ":" + os.Getenv("PORT"),
		echoTimeout:       10 * time.Second,
		readTimeout:       time.Hour,
		writeTimeout:      time.Hour,
		idleTimeout:       time.Minute,
		requestLimitation: 20,
		jwtSecret:         []byte(os.Getenv("JWT_SECRET")),
		adminWL:           strings.Split(os.Getenv("ADMIN_WHITELIST"), ","), //   "::1" means localhost!
	}
	api := &api{
		cfg:   cfg,
		store: store.NewStorage(db),
	}

	log.Fatal(api.Run())
}

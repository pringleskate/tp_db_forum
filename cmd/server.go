package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/handlers"
	forumHandler "github.com/pringleskate/tp_db_forum/internal/handlers/forum"
	profileHandler "github.com/pringleskate/tp_db_forum/internal/handlers/profile"
	serviceHandler "github.com/pringleskate/tp_db_forum/internal/handlers/service"
	"io/ioutil"
	"log"
)

func main()  {
	e := echo.New()

	connectionString := "postgres://forum_user:1221@localhost/tp_forum?sslmode=disable"
	config, err := pgx.ParseURI(connectionString)
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: 2000,
		})

	if err != nil {
		fmt.Println(err)
		return
	}

	err = LoadSchemaSQL(db)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("created tables")
	forStorage := storages.NewForumStorage(db)
	thrStorage := storages.NewThreadStorage(db)
	poStorage := storages.NewPostStorage(db)
	uStorage := storages.NewUserStorage(db)

	forHandler := forumHandler.NewHandler(thrStorage, poStorage, forStorage, uStorage)
	profHandler := profileHandler.NewHandler(uStorage)
	servHandler := serviceHandler.NewHandler(forStorage)

	handlers.Router(e, forHandler, profHandler, servHandler)

	e.Logger.Fatal(e.Start(":5000"))
}

const dbSchema = "init.sql"

func LoadSchemaSQL(db *pgx.ConnPool) error {

	content, err := ioutil.ReadFile(dbSchema)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(content)); err != nil {
		return err
	}
	tx.Commit()
	return nil
}
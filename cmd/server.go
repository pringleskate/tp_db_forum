package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/handlers"
	forumHandler "github.com/pringleskate/tp_db_forum/internal/handlers/forum"
	profileHandler "github.com/pringleskate/tp_db_forum/internal/handlers/profile"
	serviceHandler "github.com/pringleskate/tp_db_forum/internal/handlers/service"
	"github.com/pringleskate/tp_db_forum/internal/storages/forum"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
	"github.com/pringleskate/tp_db_forum/internal/storages/service"
	"io/ioutil"
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

	/*err = LoadSchemaSQL(db)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Created tables")*/
	forStorage := forum.Storage{Db: db}
	uStorage := profile.Storage{Db: db}
	servStorage := service.Service{Db: db}

	forHandler := forumHandler.NewHandler(forStorage, uStorage)
	profHandler := profileHandler.NewHandler(uStorage)
	servHandler := serviceHandler.NewHandler(servStorage)

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
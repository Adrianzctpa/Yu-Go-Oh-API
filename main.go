package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	dbConfig "Yu-Go-Oh-API/gopostgres/dbconfig"
	dbpaginate "Yu-Go-Oh-API/gopostgres/dbpaginate"
	dbUtils "Yu-Go-Oh-API/gopostgres/dbutils"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var DataSourceName = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	// Setting up the database connection
	fmt.Printf("Accessing %s ... ", os.Getenv("DB_NAME"))

	DB, err := sql.Open(dbConfig.PostgresDriver, DataSourceName)

	checkErr(err)
	defer DB.Close()

	println("Connected to database")

	app := fiber.New()

	app.Get("/cards/", func(c *fiber.Ctx) error {
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || (page <= 0) {
			page = 1
		}

		qSize, err := strconv.Atoi(c.Query("query_size"))
		if err != nil || (qSize <= 0) || (qSize > 20) {
			qSize = 20
		}

		slice := dbUtils.GetCardsInDB(DB, "", page, qSize)
		count := dbUtils.GetCount(DB, "") - 1
		pag := dbpaginate.Paginate(slice, page, qSize, count)
		return c.JSON(pag)
	})

	app.Get("/cards/load", func(c *fiber.Ctx) error {
		dbUtils.ExportJSONToDB(DB)
		return c.SendString("Done")
	})

	app.Get("/cards/filter/", func(c *fiber.Ctx) error {
		name := c.Query("name")

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || (page <= 0) {
			page = 1
		}

		qSize, err := strconv.Atoi(c.Query("query_size"))
		if err != nil || (qSize <= 0) || (qSize > 20) {
			qSize = 20
		}

		slice := dbUtils.GetCardsInDB(DB, name, page, qSize)
		count := dbUtils.GetCount(DB, name) - 1
		pag := dbpaginate.Paginate(slice, page, qSize, count)
		return c.JSON(pag)
	})

	log.Fatal(app.Listen(":4000"))
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

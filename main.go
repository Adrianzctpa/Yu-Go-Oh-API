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

		filterMap := map[string]string{
			"card_name":  "",
			"card_level": "",
		}

		json := map[string]interface{}{}

		slice, err := dbUtils.GetCardsInDB(DB, filterMap, page, qSize, "get")

		if err != nil {
			json["status"] = 500
			json["error"] = err.Error()
			return c.JSON(json)
		}

		count, url := dbUtils.GetCount(DB, filterMap, "get")
		pag := dbpaginate.Paginate(slice, page, qSize, count, url)

		json["status"] = 200
		json["data"] = pag

		return c.JSON(json)
	})

	app.Get("/cards/load", func(c *fiber.Ctx) error {
		err := dbUtils.ExportJSONToDB(DB)

		if err != nil {
			return c.JSON(fiber.Map{
				"status":  500,
				"message": "Error loading cards",
			})
		}
		return c.SendString("Done")
	})

	app.Get("/cards/filter/", func(c *fiber.Ctx) error {
		name := c.Query("card_name")
		level := c.Query("card_level")
		archetype := c.Query("archetype")
		attribute := c.Query("attribute")
		cardType := c.Query("card_type")
		race := c.Query("race")
		linkval := c.Query("linkval")
		linkmark := c.Query("linkmarkers")
		scale := c.Query("card_scale")
		atk := c.Query("atk")
		def := c.Query("def")

		arrayOfParams := []string{name, level, archetype, attribute, cardType, race, linkval, linkmark, scale, atk, def}

		noFilter := true
		for i := 0; i < len(arrayOfParams); i++ {
			if arrayOfParams[i] != "" {
				noFilter = false
			}

			if i == (len(arrayOfParams) - 1) {
				if noFilter {
					return c.SendString("No filters applied")
				}
			}
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || (page < 0) {
			page = 0
		}

		qSize, err := strconv.Atoi(c.Query("query_size"))
		if err != nil || (qSize <= 0) || (qSize > 20) {
			qSize = 20
		}

		filterMap := map[string]string{
			"card_name":   name,
			"card_level":  level,
			"archetype":   archetype,
			"attribute":   attribute,
			"card_type":   cardType,
			"race":        race,
			"linkval":     linkval,
			"linkmarkers": linkmark,
			"card_scale":  scale,
			"atk":         atk,
			"def":         def,
		}

		json := map[string]interface{}{}
		slice, err := dbUtils.GetCardsInDB(DB, filterMap, page, qSize, "filter")

		if err != nil {
			json["status"] = 500
			json["error"] = err.Error()
			return c.JSON(json)
		}

		count, url := dbUtils.GetCount(DB, filterMap, "filter")
		pag := dbpaginate.Paginate(slice, page, qSize, count, url)

		json["status"] = 200
		json["data"] = pag

		return c.JSON(json)
	})

	log.Fatal(app.Listen(":4000"))
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	dbConfig "Yu-Go-Oh-API/gopostgres/dbconfig"

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

	cards := []dbConfig.Card{}

	app.Get("/", func(c *fiber.Ctx) error {
		//exportJSONToDB(DB)
		//return c.SendString("Done")

		slice := getCardsInDB(DB, "")
		return c.JSON(slice)
	})

	app.Get("/cards/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		slice := getCardsInDB(DB, name)

		return c.JSON(slice)
	})

	app.Post("/cards", func(c *fiber.Ctx) error {
		card := new(dbConfig.Card)

		if err := c.BodyParser(card); err != nil {
			return err
		}

		if contains(cards, *card) {
			return c.SendStatus(fiber.StatusConflict)
		}

		addCardToDB(*card, DB)
		cards = append(cards, *card)
		return c.JSON(card)
	})

	log.Fatal(app.Listen(":4000"))
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func getCardsInDB(DB *sql.DB, name string) []dbConfig.Card {
	var sqlStatement string
	if name == "" {
		sqlStatement = fmt.Sprintf("SELECT * FROM %s", os.Getenv("CARD_TABLE_NAME"))
	} else {
		filter := "'" + name + "%'"
		sqlStatement = fmt.Sprintf(`SELECT * FROM %s WHERE card_name ILIKE %s`, os.Getenv("CARD_TABLE_NAME"), filter)
	}

	query, err := DB.Query(sqlStatement)
	checkErr(err)

	newCards := []dbConfig.Card{}

	for query.Next() {

		var card dbConfig.Card

		err := query.Scan(
			&card.ID, &card.Card_Name, &card.Description, &card.Archetype, &card.Card_Type, &card.Atk,
			&card.Def, &card.Card_Level, &card.Race, &card.Attr, &card.Linkval, &card.Linkmarkers, &card.Card_Scale,
		)
		checkErr(err)

		newCards = append(newCards, card)
	}

	return newCards
}

func addCardToDB(card dbConfig.Card, DB *sql.DB) {
	sqlCardStatement := fmt.Sprintf(`INSERT INTO %s (id, card_name, card_type, description, archetype, atk, def, card_level, race, attr, linkval, linkmarkers, card_scale) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, os.Getenv("CARD_TABLE_NAME"))
	prepareExecToDB(
		sqlCardStatement, DB,
		card.ID, card.Card_Name, card.Card_Type, card.Description, card.Archetype,
		card.Atk, card.Def, card.Card_Level, card.Race, card.Attr, card.Linkval, card.Linkmarkers, card.Card_Scale,
	)

	sqlImgStatement := fmt.Sprintf(`INSERT INTO %s (id, image_url, image_url_small) VALUES ($1, $2, $3)`, os.Getenv("IMAGES_TABLE_NAME"))
	prepareExecToDB(
		sqlImgStatement, DB,
		card.ID, card.Images.Image_url, card.Images.Image_url_small,
	)
}

func contains(s []dbConfig.Card, e dbConfig.Card) bool {
	for _, a := range s {
		if a.ID == e.ID {
			return true
		}
	}
	return false
}

func prepareExecToDB(sqlStatement string, DB *sql.DB, args ...interface{}) {
	insert, err := DB.Prepare(sqlStatement)
	checkErr(err)

	result, err := insert.Exec(args...)
	checkErr(err)

	affect, err := result.RowsAffected()
	checkErr(err)

	fmt.Println(affect)
}

func exportJSONToDB(DB *sql.DB) {
	jsonFile, err := os.Open("cardinfo.json")
	checkErr(err)

	defer jsonFile.Close()

	byteVal, _ := io.ReadAll(jsonFile)

	var data dbConfig.DB

	json.Unmarshal(byteVal, &data)

	// add every card to the database
	for i := 0; i < len(data.Cards); i++ {
		fmt.Println("Adding card: ", data.Cards[i].Card_Name)
		addCardToDB(data.Cards[i], DB)
	}
}

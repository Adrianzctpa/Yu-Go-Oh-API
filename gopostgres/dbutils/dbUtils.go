package dbutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	dbConfig "Yu-Go-Oh-API/gopostgres/dbconfig"

	_ "github.com/lib/pq"
)

func GetCount(DB *sql.DB, filter string) int {
	var count int
	var sqlStatement string

	if filter == "" {
		sqlStatement = writeSQLStatement("count", "", 0, 0)
	} else {
		filter := "'" + filter + "%'"
		sqlStatement = writeSQLStatement("countFilter", filter, 0, 0)
	}

	err := DB.QueryRow(sqlStatement).Scan(&count)

	checkErr(err)

	return count
}

func GetCardsInDB(DB *sql.DB, name string, page int, query_size int) []dbConfig.Card {
	var sqlStatement string
	if name == "" {
		sqlStatement = writeSQLStatement("get", "", page, query_size)
	} else {
		filter := "'" + name + "%'"
		sqlStatement = writeSQLStatement("filter", filter, page, query_size)
	}

	query, err := DB.Query(sqlStatement)
	checkErr(err)

	newCards := []dbConfig.Card{}

	for query.Next() {

		var card dbConfig.Card
		var arr []string
		var small_arr []string

		err = query.Scan(
			&card.ID, &card.Card_Name, &card.Description, &card.Archetype, &card.Card_Type, &card.Atk,
			&card.Def, &card.Card_Level, &card.Race, &card.Attr, &card.Linkval, &card.Linkmarkers, &card.Card_Scale,
			&card.Image_url_uint8, &card.Image_url_small_uint8,
		)
		checkErr(err)

		parsed_image := string(card.Image_url_uint8[:])
		small_parsed_image := string(card.Image_url_small_uint8[:])

		image_slice := strings.Split(parsed_image, ",")
		small_image_slice := strings.Split(small_parsed_image, ",")

		for i := 0; i < len(image_slice); i++ {
			parsed_string := strings.ReplaceAll(image_slice[i], "{", "")
			parsed_string = strings.ReplaceAll(parsed_string, "}", "")

			arr = append(arr, parsed_string)

			small_parsed_string := strings.ReplaceAll(small_image_slice[i], "{", "")
			small_parsed_string = strings.ReplaceAll(small_parsed_string, "}", "")

			small_arr = append(small_arr, small_parsed_string)
		}

		card.Image_url = arr
		card.Image_url_small = small_arr
		newCards = append(newCards, card)
	}

	return newCards
}

func AddCardToDB(card dbConfig.CardDB, DB *sql.DB) {
	sqlCardStatement := writeSQLStatement("post", "", 0, 0)
	prepareExecToDB(
		sqlCardStatement, DB,
		card.ID, card.Card_Name, card.Card_Type, card.Description, card.Archetype,
		card.Atk, card.Def, card.Card_Level, card.Race, card.Attr, card.Linkval, card.Linkmarkers, card.Card_Scale,
	)

	sqlImgStatement := writeSQLStatement("postImg", "", 0, 0)
	if len(card.Images) > 1 {
		for i := 0; i < (len(card.Images) - 1); i++ {
			prepareExecToDB(
				sqlImgStatement, DB,
				card.Images[i].ID, card.ID, card.Images[i].Image_url, card.Images[i].Image_url_small,
			)
		}
	} else {
		prepareExecToDB(
			sqlImgStatement, DB,
			card.Images[0].ID, card.ID, card.Images[0].Image_url, card.Images[0].Image_url_small,
		)
	}
}

func prepareExecToDB(sqlStatement string, DB *sql.DB, args ...interface{}) {
	insert, err := DB.Prepare(sqlStatement)
	checkErr(err)

	result, err := insert.Exec(args...)
	checkErr(err)

	rowsAffected, err := result.RowsAffected()
	checkErr(err)

	fmt.Printf("Rows affected: %d \n", rowsAffected)
}

func writeSQLStatement(statementType string, filter string, page int, limit int) string {
	var sqlStatement string

	if page > 1 {
		page = page * 10
	}

	switch statementType {
	case "filter":
		sqlStatement = fmt.Sprintf(
			`
				SELECT *
				FROM (SELECT * FROM %s WHERE card_name ILIKE %s LIMIT %d OFFSET %d) as Q,
				LATERAL (
					SELECT array_agg(image_url::text) as image_url, array_agg(image_url_small::text) as image_url_small
					FROM %s ci
					WHERE ci.card_id = Q.id
				) as L
				`, os.Getenv("CARD_TABLE_NAME"), filter, limit, page, os.Getenv("IMAGES_TABLE_NAME"))
	case "get":
		sqlStatement = fmt.Sprintf(`
				SELECT *
				FROM (SELECT * FROM %s LIMIT %d OFFSET %d) as Q,
				LATERAL (
					SELECT array_agg(image_url::text) as image_url, array_agg(image_url_small::text) as image_url_small
					FROM %s ci
					WHERE ci.card_id = Q.id
				) as L
				`, os.Getenv("CARD_TABLE_NAME"), limit, page, os.Getenv("IMAGES_TABLE_NAME"))
	case "post":
		sqlStatement = fmt.Sprintf(`
				INSERT INTO %s 
				(id, card_name, card_type, description, archetype, atk, def, card_level, race, attr, linkval, linkmarkers, card_scale) 
				VALUES 
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, os.Getenv("CARD_TABLE_NAME"))
	case "postImg":
		sqlStatement = fmt.Sprintf(`INSERT INTO %s (id, card_id, image_url, image_url_small) VALUES ($1, $2, $3, $4)`, os.Getenv("IMAGES_TABLE_NAME"))
	case "count":
		sqlStatement = fmt.Sprintf(`SELECT COUNT(*) FROM %s`, os.Getenv("CARD_TABLE_NAME"))
	case "countFilter":
		sqlStatement = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE card_name ILIKE %s`, os.Getenv("CARD_TABLE_NAME"), filter)
	}
	return sqlStatement
}

func ExportJSONToDB(DB *sql.DB) {
	jsonFile, err := os.Open("cardinfo.json")
	checkErr(err)

	defer jsonFile.Close()

	byteVal, _ := io.ReadAll(jsonFile)

	var data dbConfig.DB

	json.Unmarshal(byteVal, &data)

	// add every card to the database
	for i := 0; i < len(data.Cards); i++ {
		AddCardToDB(data.Cards[i], DB)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

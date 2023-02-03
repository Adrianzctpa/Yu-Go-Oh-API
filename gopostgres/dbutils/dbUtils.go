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

func GetCount(DB *sql.DB, filterMap map[string]string, mode string) (int, string) {
	var count int
	var sqlStatement string
	var url string

	if mode == "filter" {
		sqlStatement, url = writeSQLStatement("countFilter", filterMap, 0, 0)
	} else {
		sqlStatement, url = writeSQLStatement("count", filterMap, 0, 0)
	}

	err := DB.QueryRow(sqlStatement).Scan(&count)

	checkErr(err)

	return count, url
}

func cleanStringAndReturnArr(str string) []string {
	slice := strings.Split(str, ",")
	var arr []string
	for i := 0; i < len(slice); i++ {
		parsed_string := strings.ReplaceAll(slice[i], "{", "")
		parsed_string = strings.ReplaceAll(parsed_string, "}", "")

		slice[i] = parsed_string

		arr = append(arr, parsed_string)
	}

	return arr
}

func GetCardById(DB *sql.DB, id int) (dbConfig.Card, error) {
	var sqlStatement string
	filterMap := map[string]string{"id": fmt.Sprintf("%d", id)}

	sqlStatement, _ = writeSQLStatement("getById", filterMap, 0, 0)

	query, err := DB.Query(sqlStatement)

	if checkErr(err) {
		return dbConfig.Card{}, err
	}

	var card dbConfig.Card

	for query.Next() {
		err = query.Scan(
			&card.ID, &card.Card_Name, &card.Card_Type, &card.Description, &card.Archetype, &card.Atk,
			&card.Def, &card.Card_Level, &card.Race, &card.Attr, &card.Linkval, &card.Linkmarkers, &card.Card_Scale,
			&card.Image_url_uint8, &card.Image_url_small_uint8,
		)

		if checkErr(err) {
			return dbConfig.Card{}, err
		}

		parsed_image := string(card.Image_url_uint8[:])
		small_parsed_image := string(card.Image_url_small_uint8[:])

		arr := cleanStringAndReturnArr(parsed_image)
		small_arr := cleanStringAndReturnArr(small_parsed_image)

		card.Image_url = arr
		card.Image_url_small = small_arr
	}

	return card, err
}

func GetCardsInDB(DB *sql.DB, filterArr map[string]string, page int, query_size int, mode string) ([]dbConfig.Card, error) {
	var sqlStatement string

	if mode == "filter" {
		sqlStatement, _ = writeSQLStatement("filter", filterArr, page, query_size)
	} else {
		sqlStatement, _ = writeSQLStatement("get", filterArr, page, query_size)
	}

	query, err := DB.Query(sqlStatement)

	if checkErr(err) {
		return []dbConfig.Card{}, err
	}

	newCards := []dbConfig.Card{}

	for query.Next() {

		var card dbConfig.Card
		var arr []string
		var small_arr []string

		err = query.Scan(
			&card.ID, &card.Card_Name, &card.Card_Type, &card.Description, &card.Archetype, &card.Atk,
			&card.Def, &card.Card_Level, &card.Race, &card.Attr, &card.Linkval, &card.Linkmarkers, &card.Card_Scale,
			&card.Image_url_uint8, &card.Image_url_small_uint8,
		)
		checkErr(err)

		if checkErr(err) {
			return []dbConfig.Card{}, err
		}

		parsed_image := string(card.Image_url_uint8[:])
		small_parsed_image := string(card.Image_url_small_uint8[:])

		arr = cleanStringAndReturnArr(parsed_image)
		small_arr = cleanStringAndReturnArr(small_parsed_image)

		card.Image_url = arr
		card.Image_url_small = small_arr
		newCards = append(newCards, card)
	}

	return newCards, err
}

func AddCardToDB(card dbConfig.CardDB, DB *sql.DB) {
	filterMap := map[string]string{
		"card_name":  "",
		"card_level": "",
	}

	sqlCardStatement, _ := writeSQLStatement("post", filterMap, 0, 0)
	prepareExecToDB(
		sqlCardStatement, DB,
		card.ID, card.Card_Name, card.Card_Type, card.Description, card.Archetype,
		card.Atk, card.Def, card.Card_Level, card.Race, card.Attr, card.Linkval, card.Linkmarkers, card.Card_Scale,
	)

	sqlImgStatement, _ := writeSQLStatement("postImg", filterMap, 0, 0)
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

func AddBanlistToDB(banlist dbConfig.Banlist, DB *sql.DB, mode string) {
	filterMap := map[string]string{
		"table": os.Getenv("BANLIST_TABLE"),
	}

	if mode == "ocg" {
		filterMap["table"] = os.Getenv("OCG_BANLIST_TABLE")
	}

	jsonStr, err := json.Marshal(banlist.BanlistInfo)
	checkErr(err)

	sqlStatement, _ := writeSQLStatement("postBanlist", filterMap, 0, 0)

	prepareExecToDB(
		sqlStatement, DB,
		banlist.ID, banlist.ID, jsonStr, banlist.FrameType,
	)
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

func writeSQLStatement(statementType string, filterMap map[string]string, page int, limit int) (string, string) {
	baseUrl := "/cards/?"

	if page > 1 {
		page = (page - 1) * limit
	} else {
		page = 0
	}

	switch statementType {
	case "filter":
		sqlStatement, url := filterLoop(filterMap, limit, page, "filter")

		return sqlStatement, url
	case "get":
		sqlStatement := fmt.Sprintf(`
				SELECT *
				FROM (SELECT * FROM %s LIMIT %d OFFSET %d) as Q,
				LATERAL (
					SELECT array_agg(image_url::text) as image_url, array_agg(image_url_small::text) as image_url_small
					FROM %s ci
					WHERE ci.card_id = Q.id
				) as L
				`, os.Getenv("CARD_TABLE_NAME"), limit, page, os.Getenv("IMAGES_TABLE_NAME"))

		return sqlStatement, baseUrl
	case "getById":
		sqlStatement := fmt.Sprintf(`
		SELECT *
		FROM (SELECT * FROM %s WHERE id = %s) as Q,
		LATERAL (
			SELECT array_agg(image_url::text) as image_url, array_agg(image_url_small::text) as image_url_small
			FROM %s ci
			WHERE ci.card_id = Q.id
		) as L
		`, os.Getenv("CARD_TABLE_NAME"), filterMap["id"], os.Getenv("IMAGES_TABLE_NAME"))

		return sqlStatement, baseUrl
	case "post":
		sqlStatement := fmt.Sprintf(`
				INSERT INTO %s 
				(id, card_name, card_type, description, archetype, atk, def, card_level, race, attr, linkval, linkmarkers, card_scale) 
				VALUES 
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, os.Getenv("CARD_TABLE_NAME"))

		return sqlStatement, baseUrl
	case "postImg":
		sqlStatement := fmt.Sprintf(`INSERT INTO %s (id, card_id, image_url, image_url_small) VALUES ($1, $2, $3, $4)`, filterMap["table"])

		return sqlStatement, baseUrl
	case "postBanlist":
		sqlStatement := fmt.Sprintf(`INSERT INTO %s (id, card_id, banlist_info, frameType) VALUES ($1, $2, $3, $4)`, os.Getenv("BANLIST_TABLE_NAME"))

		return sqlStatement, baseUrl
	case "count":
		sqlStatement := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, os.Getenv("CARD_TABLE_NAME"))

		return sqlStatement, baseUrl
	case "countFilter":
		sqlStatement, url := filterLoop(filterMap, limit, page, "count")

		return sqlStatement, url
	}

	return "", ""
}

func filterLoop(filterMap map[string]string, limit int, page int, mode string) (string, string) {
	filterUrl := `/cards/filter/?`
	sqlStatement := fmt.Sprintf(`
	SELECT *
	FROM (SELECT * FROM %s WHERE
	`, os.Getenv("CARD_TABLE_NAME"))

	if mode == "count" {
		sqlStatement = fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE 
		`, os.Getenv("CARD_TABLE_NAME"))
	}

	i := 0
	edited := false

	for key, value := range filterMap {
		i++

		if value != "" {
			// Filtering exacts
			arrayOfExacts := [6]string{"card_level", "card_type", "linkval", "card_scale", "atk", "def"}
			for i := 0; i < len(arrayOfExacts); i++ {
				if key == arrayOfExacts[i] {
					var newFilter string
					if edited {
						newFilter = fmt.Sprintf("AND %s = '%s'", key, value)
						filterUrl = filterUrl + fmt.Sprintf(`&%s=%s`, key, value)
					} else {
						newFilter = fmt.Sprintf("%s = '%s'", key, value)
						filterUrl = filterUrl + fmt.Sprintf(`%s=%s`, key, value)
						edited = true
					}
					sqlStatement = sqlStatement + newFilter
				}
			}

			switch key {
			case "card_name":
				value = strings.ReplaceAll(value, `"`, "")
				filter := "'" + value + "%'"

				if edited {
					sqlStatement = sqlStatement + fmt.Sprintf(`
									AND card_name ILIKE %s OR description ILIKE %s
									`, filter, filter)
					filterUrl = filterUrl + fmt.Sprintf(`&%s=%s`, key, value)
				} else {
					sqlStatement = sqlStatement + fmt.Sprintf(`
									card_name ILIKE %s OR description ILIKE %s
									`, filter, filter)
					filterUrl = filterUrl + fmt.Sprintf(`%s=%s`, key, value)
					edited = true
				}

			case "linkmarkers":
				var newFilter string

				linkmark := strings.ReplaceAll(value, `"`, `'`)
				newFilter = fmt.Sprintf("AND %s @> ARRAY[%s]::text[]", key, linkmark)

				if edited {
					filterUrl = filterUrl + fmt.Sprintf(`&%s=%s`, key, value)
				} else {
					newFilter = fmt.Sprintf("%s @> ARRAY[%s]::text[]", key, linkmark)
					filterUrl = filterUrl + fmt.Sprintf(`%s=%s`, key, value)
					edited = true
				}

				sqlStatement = sqlStatement + newFilter
			default:
				exists := false
				for _, val := range arrayOfExacts {
					if key == val {
						exists = true
					}
				}

				if !exists {
					value = strings.ReplaceAll(value, `"`, "")
					filter := "'" + value + "%'"

					if edited {
						sqlStatement = sqlStatement + fmt.Sprintf(`
										AND %s ILIKE %s
										`, key, filter)
						filterUrl = filterUrl + fmt.Sprintf(`&%s=%s`, key, value)
					} else {
						sqlStatement = sqlStatement + fmt.Sprintf(`
										%s ILIKE %s
										`, key, filter)
						filterUrl = filterUrl + fmt.Sprintf(`%s=%s`, key, value)
						edited = true
					}
				}
			}
		}

		// If it is the end of the loop
		if i == len(filterMap) {
			if mode != "count" {
				sqlStatement = sqlStatement + fmt.Sprintf(` 
				LIMIT %d OFFSET %d) as Q, 
				LATERAL (
					SELECT array_agg(image_url::text) as image_url, array_agg(image_url_small::text) as image_url_small 
					FROM %s ci 
					WHERE ci.card_id = Q.id
				) as L`, limit, page, os.Getenv("IMAGES_TABLE_NAME"))
			}

			filterUrl = filterUrl + "&"
		}
	}

	return sqlStatement, filterUrl
}

func ExportJSONToDB(DB *sql.DB) error {
	jsonFile, err := os.Open("cardinfo.json")
	if checkErr(err) {
		return err
	}

	defer jsonFile.Close()

	byteVal, _ := io.ReadAll(jsonFile)

	var data dbConfig.DB

	json.Unmarshal(byteVal, &data)

	// add every card to the database
	for i := 0; i < len(data.Cards); i++ {
		AddCardToDB(data.Cards[i], DB)
	}

	return nil
}

func ExportBanlistJSONToDB(DB *sql.DB, mode string) error {
	var jsonFile *os.File
	var err error

	if mode == "tcg" {
		jsonFile, err = os.Open("banlist.json")
	} else {
		jsonFile, err = os.Open("banlistocg.json")
	}

	if checkErr(err) {
		return err
	}

	defer jsonFile.Close()

	byteVal, _ := io.ReadAll(jsonFile)

	var data dbConfig.BanlistDB

	json.Unmarshal(byteVal, &data)

	for i := 0; i < len(data.List); i++ {
		AddBanlistToDB(data.List[i], DB, mode)
	}

	return nil
}

func checkErr(err error) bool {
	return err != nil
}

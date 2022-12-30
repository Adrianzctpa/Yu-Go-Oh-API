package dbconfig

import (
	pq "github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
)

type DB struct {
	Cards []CardDB `json:"data"`
}

type Card struct {
	ID                    int            `json:"id"`
	Card_Name             string         `json:"card_name"`
	Card_Type             string         `json:"card_type"`
	Description           string         `json:"description"`
	Archetype             string         `json:"archetype"`
	Atk                   null.Int       `json:"atk"`
	Def                   null.Int       `json:"def"`
	Card_Level            null.Int       `json:"card_level"`
	Race                  null.String    `json:"race"`
	Attr                  null.String    `json:"attribute"`
	Linkval               null.Int       `json:"linkval"`
	Linkmarkers           pq.StringArray `json:"linkmarkers"`
	Card_Scale            null.Int       `json:"card_scale"`
	Image_url_uint8       []byte
	Image_url_small_uint8 []byte
	Image_url             []string `json:"image_url"`
	Image_url_small       []string `json:"image_url_small"`
}

type CardDB struct {
	ID          int            `json:"id"`
	Card_Name   string         `json:"card_name"`
	Card_Type   string         `json:"card_type"`
	Description string         `json:"description"`
	Archetype   string         `json:"archetype"`
	Atk         null.Int       `json:"atk"`
	Def         null.Int       `json:"def"`
	Card_Level  null.Int       `json:"card_level"`
	Race        null.String    `json:"race"`
	Attr        null.String    `json:"attribute"`
	Linkval     null.Int       `json:"linkval"`
	Linkmarkers pq.StringArray `json:"linkmarkers"`
	Card_Scale  null.Int       `json:"card_scale"`
	Images      []CardImageDB  `json:"card_images"`
}

type CardImageDB struct {
	ID              int    `json:"id"`
	Image_url       string `json:"image_url"`
	Image_url_small string `json:"image_url_small"`
}

const PostgresDriver = "postgres"

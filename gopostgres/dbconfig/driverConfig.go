package dbconfig

import pq "github.com/lib/pq"

type DB struct {
	Cards []CardDB `json:"data"`
}

type Card struct {
	ID                    int            `json:"id"`
	Card_Name             string         `json:"card_name"`
	Card_Type             string         `json:"card_type"`
	Description           string         `json:"description"`
	Archetype             string         `json:"archetype"`
	Atk                   int            `json:"atk"`
	Def                   int            `json:"def"`
	Card_Level            int            `json:"card_level"`
	Race                  string         `json:"race"`
	Attr                  string         `json:"attribute"`
	Linkval               int            `json:"linkval"`
	Linkmarkers           pq.StringArray `json:"linkmarkers"`
	Card_Scale            int            `json:"card_scale"`
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
	Atk         int            `json:"atk"`
	Def         int            `json:"def"`
	Card_Level  int            `json:"card_level"`
	Race        string         `json:"race"`
	Attr        string         `json:"attribute"`
	Linkval     int            `json:"linkval"`
	Linkmarkers pq.StringArray `json:"linkmarkers"`
	Card_Scale  int            `json:"card_scale"`
	Images      []CardImageDB  `json:"card_images"`
}

type CardImageDB struct {
	ID              int    `json:"id"`
	Image_url       string `json:"image_url"`
	Image_url_small string `json:"image_url_small"`
}

func toJson(card CardDB) map[string]any {
	json := map[string]any{
		"id":          card.ID,
		"card_name":   card.Card_Name,
		"type":        card.Card_Type,
		"description": card.Description,
		"archetype":   card.Archetype,
		"atk":         card.Atk,
		"def":         card.Def,
		"card_level":  card.Card_Level,
		"race":        card.Race,
		"attr":        card.Attr,
		"linkval":     card.Linkval,
		"linkmarkers": card.Linkmarkers,
		"card_scale":  card.Card_Scale,
		"images":      card.Images,
	}
	return json
}

func fromJson(json map[string]any) CardDB {
	card := CardDB{
		ID:          json["id"].(int),
		Card_Name:   json["card_name"].(string),
		Card_Type:   json["type"].(string),
		Description: json["description"].(string),
		Archetype:   json["archetype"].(string),
		Atk:         json["atk"].(int),
		Def:         json["def"].(int),
		Card_Level:  json["card_level"].(int),
		Race:        json["race"].(string),
		Attr:        json["attr"].(string),
		Linkval:     json["linkval"].(int),
		Linkmarkers: json["linkmarkers"].(pq.StringArray),
		Card_Scale:  json["card_scale"].(int),
		Images:      json["images"].([]CardImageDB),
	}
	return card
}

const PostgresDriver = "postgres"

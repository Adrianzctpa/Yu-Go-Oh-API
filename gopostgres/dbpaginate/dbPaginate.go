package dbpaginate

import (
	"fmt"
	"math"

	dbConfig "Yu-Go-Oh-API/gopostgres/dbconfig"

	_ "github.com/lib/pq"
)

func Paginate(query []dbConfig.Card, page int, qSize int, count int) map[string]interface{} {
	pages := math.Ceil((float64(count) / float64(qSize)))

	response := map[string]interface{}{
		"cards":      query,
		"pages":      pages,
		"total":      count,
		"query_size": qSize,
	}

	start := (page - 1) * qSize
	end := start + qSize

	if end >= count {
		response["next"] = nil
	} else {
		response["next"] = fmt.Sprintf("/cards?page=%d&query_size=%d", page+1, qSize)
	}

	if page > 1 {
		response["prev"] = fmt.Sprintf("/cards?page=%d&query_size=%d", page-1, qSize)
	} else {
		response["prev"] = nil
	}

	return response
}

package authorization

import (
	"context"
	"log"
	"redminetb/internal/data"
)

func FetchUsernamesFromDB(tgacc string) string {
	var nameSurname string
	db := data.GetDB()
	err := db.QueryRow(context.Background(), "SELECT namesurname FROM tgacc WHERE tgacc = $1", tgacc).Scan(&nameSurname)
	if err != nil {
		log.Println("Query failed:", err)
		return ""
	}
	return nameSurname
}

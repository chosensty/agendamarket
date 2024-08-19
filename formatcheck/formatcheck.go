package formatcheck

import (
	"database/sql"
  "strconv"
  db "agendamarket/database"
  _ "github.com/go-sql-driver/mysql"

)

func CheckFormat(keyword string, args []string, database *sql.DB) (string, bool) {
  if keyword == "buy" || keyword == "sell"{
    // checking for the character number of keywords
    if len(args) < 4 {
      return "Please input the correct format", false
    }

    // checking if the stock exists
    if !db.StockExists(database, args[2]) {
      return "Please enter a valid stock name", false
    }

    var value float64
    var err error
    first_char := args[3][0]
    // checking if the stock number of price given is valid
    if string(first_char) == "$" {
      value, err = strconv.ParseFloat(args[3][1:], 64)
      if err != nil {
        return "Please input the correct format", false
      }
    } else {
      value, err = strconv.ParseFloat(args[3], 64)
      if err != nil {
        return "Please input the correct format", false
      }
    }
    if value < 0.1 {
      return "You cannot " + keyword + " less than 0.1 shares", false
    }
    return "", true
  }
  return "", true

}

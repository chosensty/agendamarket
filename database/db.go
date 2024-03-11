package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func GetStockPrice(db *sql.DB, stock_name string) float64 {
	query := `SELECT price FROM Stock WHERE name=?`
	var price float64
	err := db.QueryRow(query, stock_name).Scan(&price)
	if err != nil {
		return -1.0
	}
	return price
}

func ReturnStock(db *sql.DB, input string) string {
	query := `SELECT name, price, total_shares FROM Stock`

	var name string
	var price string
	var total_shares string
	response := ""
	if input == "*" || input == "all" {
		rows, err := db.Query(query)

		if err != nil {
			log.Println(err)
		}
		for rows.Next() {
			err := rows.Scan(&name, &price, &total_shares)
			if err != nil {
				log.Println(err)
			}
			response += name + " stock is currently worth $" + price + " per share.\n"
		}
	} else if StockExists(db, input) {
		price := GetStockPrice(db, input)
		price_string := strconv.FormatFloat(price, 'f', 2, 64)
		response = input + " stock is currently worth $" + price_string + " per share.\n"
	} else {
		response = "Could not find stock \"" + input + "\""
	}
	return response
}
func NewUser(db *sql.DB, ID string) bool {
	query := "INSERT INTO User VALUES (1000, ?)"
	fmt.Println(query)
	_, err := db.Exec(query, ID)

	if err != nil {
		log.Print(err)
		return true
	}
	return false
}
func GetUserShares(db *sql.DB, userID string, stock_name string) float64 {
	query := `SELECT shares FROM UserStocks WHERE userID=? AND stockName=?`
	var shares float64
	err := db.QueryRow(query, userID, stock_name).Scan(&shares)
	if err != nil {
		if err == sql.ErrNoRows {
			query = `INSERT INTO UserStocks (userID, stockName, shares) VALUES (?, ?, ?)`
			_, err2 := db.Exec(query, userID, stock_name, 0.0)
			if err2 != nil {
				return -1.0
			} else {
				return 0.0
			}
		}
		log.Print(err)
		return -1.0
	}
	return shares
}
func RemoveUserShares(db *sql.DB, userID string, stock_name string) bool {
	query := `DELETE FROM UserStock WHERE stockName=? AND userID=?`
	_, err := db.Exec(query, stock_name, userID)
	return err == nil
}

func PreciseAdd(n1 float64, n2 float64) float64 {
	num1 := decimal.NewFromFloat(n1)
	num2 := decimal.NewFromFloat(n2)
	result := num1.Add(num2)
	output, _ := result.Float64()
	return output
}

func PreciseMult(n1 float64, n2 float64) float64 {
	num1 := decimal.NewFromFloat(n1)
	num2 := decimal.NewFromFloat(n2)
	result := num1.Mul(num2)
	output, _ := result.Float64()
	return output
}

func PreciseDiv(n1 float64, n2 float64) float64 {
	num1 := decimal.NewFromFloat(n1)
	num2 := decimal.NewFromFloat(n2)
	result := num1.Div(num2)
	output, _ := result.Float64()
	return output
}

func PreciseSub(n1 float64, n2 float64) float64 {
	num1 := decimal.NewFromFloat(n1)
	num2 := decimal.NewFromFloat(n2)
	result := num1.Sub(num2)
	output, _ := result.Float64()
	return output
}
func RoundFloat(value float64, dp int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded_dec, _ := dec.Round(dp).Float64()
	return rounded_dec
}
func StockTransaction(db *sql.DB, userID string, name string, price string, sign int, tax float64) (bool, string) {
	balance := BalCheck(db, userID)
	if balance == -1.0 {
		return false, "You have not registered an account."
	}

	price_float, _ := strconv.ParseFloat(price, 64)

	converted_sign := float64(sign)

	bal_change := PreciseMult(price_float, converted_sign)

	bal_change = bal_change / (1 - tax)

	new_balance := PreciseSub(balance, bal_change)

	if RoundFloat(new_balance, 2) < 0 {
		return false, "You do not have enough funds to complete this purchase."
	}

	new_balance = RoundFloat(new_balance, 2)

	var query string

	orig_price := GetStockPrice(db, name)
	if orig_price == -1.0 {
		return false, "Could not find stock " + name
	}

	sens, _ := strconv.ParseFloat(os.Getenv("global_sensitivity"), 64)

	share_num := PreciseDiv(price_float, orig_price)

	user_shares := GetUserShares(db, userID, name)
	if user_shares == -1.0 {
		return false, "An error has occured while trying to obtain your share number."
	}

	value := user_shares - share_num

	if sign == -1 && RoundFloat(value, 2) < 0 {
		return false, "You do not have enough shares to sell"
	}

	user_shares = PreciseAdd(user_shares, PreciseMult(share_num, converted_sign))

	if RoundFloat(user_shares, 2) > 10 {
		return false, "You can not own more than 10 shares."
	}

	query = `UPDATE UserStocks SET shares=? WHERE userID=? AND stockName=?`

	_, err := db.Exec(query, user_shares, userID, name)
	if RoundFloat(value, 2) == 0 {
		RemoveUserShares(db, userID, name)
	}
	if err != nil {
		log.Print(err)
		return false, "Error occured while updating your shares."
	}

	query = `UPDATE User SET balance=? WHERE userID=?`

	_, err = db.Exec(query, new_balance, userID)
	if err != nil {
		log.Println(err)
	}

	stock_price := PreciseMult(orig_price, (1 + PreciseMult(sens, share_num*converted_sign)))

	query = "UPDATE Stock SET price=? WHERE name=?"
	_, err = db.Exec(query, stock_price, name)

	if err != nil {
		log.Print(err)
		return false, "Error occured."
	}

	query = `INSERT INTO Transactions (userID, stockName, price, stockNum, type, stockPrice) VALUES (?, ?, ?, ?, ?, ?)`
	var buysell string
	if sign == 1 {
		buysell = "b"

	} else {
		buysell = "s"
	}

	_, err = db.Exec(query, userID, name, price_float, share_num, buysell, stock_price)

	if err != nil {
		log.Print(err)
	}

	share_string := strconv.FormatFloat(share_num, 'f', 2, 64)

	if sign == 1 {
		return true, "Bought " + share_string + " " + name + " shares"
	} else {
		return true, "Sold " + share_string + " " + name + " shares"
	}

}
func BalCheck(db *sql.DB, ID string) float64 {
	query := "SELECT balance FROM User WHERE userID=?"
	var o float64
	err := db.QueryRow(query, ID).Scan(&o)
	if err != nil {
		log.Print(err)
		return -1.0
	}
	return o
}

func GetNetWorth(db *sql.DB, ID string) string {
	balance := BalCheck(db, ID)
	output := ""
	query := "SELECT stockName, shares FROM UserStocks WHERE userID=?"
	data, err := db.Query(query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return output
		}
		return "Error occured."
	}
	for data.Next() {
		var name string
		var shares float64
		data.Scan(&name, &shares)
		query = "SELECT price FROM Stock WHERE name=?"
		var price float64
		db.QueryRow(query, name).Scan(&price)
		balance += price * shares
	}
	output = "Your current net worth is $" + strconv.FormatFloat(balance, 'f', 2, 64)
	return output
}
func UserList(db *sql.DB, userID string) string {
	query := "SELECT stockName, shares FROM UserStocks WHERE userID=?"
	data, err := db.Query(query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "You have not invested into any stocks."
		} else {
			return "Error occured."
		}
	}
	defer data.Close()

	var response string = ""
	for data.Next() {
		var name string
		var shares float64
		err := data.Scan(&name, &shares)
		if err != nil {
			fmt.Println(err)
		}
		response += "You own " + strconv.FormatFloat(shares, 'f', 2, 64) + " " + name + " stocks\n"
		fmt.Println(name)
	}

	return response
}

func StockExists(db *sql.DB, stockName string) bool {
	query := "SELECT name FROM Stock WHERE name=?"
	var o string
	err := db.QueryRow(query, stockName).Scan(&o)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print("User tried to find a stock that does not exist.")
		} else {
			log.Print(err)
		}
		return false
	}
	return true
}

func UserExists(db *sql.DB, input string) bool {
	query := "SELECT userID FROM User WHERE userID=" + input
	var o string
	err := db.QueryRow(query).Scan(&o)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return false
	}
	if o == input {
		return true
	}
	return false
}
func FindRow(db *sql.DB, input string) {

	query := "SELECT name, price, total_shares FROM Stock WHERE name = \"" + input + "\""

	var name string
	var price float64
	var total_shares int
	err := db.QueryRow(query).Scan(&name, &price, &total_shares)
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("%s %g %d\n", name, price, total_shares)
}

func RemoveStock(db *sql.DB, stock_name string) bool {
	query := `DELETE FROM Stock WHERE name=?`
	_, err := db.Exec(query, stock_name)
	if err != nil {
		log.Print(err)
		return false
	}
	query = `DELETE FROM Transactions WHERE stockName=?`
	_, err = db.Exec(query, stock_name)
	if err != nil {
		log.Print(err)
		return false
	}
	query = `DELETE FROM UserStocks WHERE stockName=?`
	_, err = db.Exec(query, stock_name)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}
func NewStock(db *sql.DB, stock_name string, base_price float64) bool {
	query := "INSERT INTO Stock (name, price, total_shares) VALUES (?, ?, ?)"
	_, err := db.Exec(query, stock_name, base_price, 10000)
	if err != nil {
		log.Println("Couldn't create new stock.")
	}

	return true
}

func Init() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("couldn't load .env file")
	}
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	log.Println("Successfully connected to PlanetScale!")
	return db
}

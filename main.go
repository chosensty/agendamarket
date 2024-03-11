package main

import (
	"encoding/json"
	"fmt"
	db "go-discord-bot/database"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const prefix string = "!jdl"

type Help struct {
	Command     string
	Format      string
	Explanation string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("couldn't load .env file")
	}
	sess, err := discordgo.New("Bot " + os.Getenv(("TOKEN")))
	if err != nil {
		log.Fatal(err)
	}
	database := db.Init()

	//	db.FindRow(database, "luffy")
	helpFileBytes, err := os.ReadFile("./src/help.json")
	if err != nil {
		log.Fatal(err)
	}

	var helpSlice []Help

	err = json.Unmarshal(helpFileBytes, &helpSlice)

	if err != nil {
		log.Fatal(err)
	}

	helpMessage := ""

	for _, e := range helpSlice {
		helpMessage += e.Command + ": " + e.Format + "\n"
	}
	helpMessage += "A deeper explanation of each command can be accessed using !jdl stock help <command name>"

	yes_button := discordgo.Button{
		Label:    "Yes",
		Style:    discordgo.SuccessButton,
		CustomID: "yes_button",
	}

	no_button := discordgo.Button{
		Label:    "No",
		Style:    discordgo.DangerButton,
		CustomID: "no_button",
	}
	yn_button_row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{yes_button, no_button},
	}
	tax, _ := strconv.ParseFloat(os.Getenv("TAX"), 64)
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		fmt.Println(m.Content)

		args := strings.Split(m.Content, " ")

		if args[0] != prefix {
			return
		}
		if args[1] == "stock" {
			if len(args) > 2 {
				switch args[2] {
				case "new":
					if m.Author.ID == os.Getenv("ADMIN") && len(args) > 4 {
						price, err := strconv.ParseFloat(args[4], 64)
						if err == nil {
							cond := db.NewStock(database, args[3], price)
							if cond {
								s.ChannelMessageSend(m.ChannelID, "Add new stock "+args[3]+" with a base price of "+args[4])
							}
						}
					}
				case "remove":
					if m.Author.ID == os.Getenv("ADMIN") && len(args) > 3 {
						if db.RemoveStock(database, args[3]) {
							s.ChannelMessageSend(m.ChannelID, "Remove stock "+args[3])
						}
					}
				case "balance":
					bal := db.BalCheck(database, m.Author.ID)
					s.ChannelMessageSend(m.ChannelID, "Your current balance is "+strconv.FormatFloat(bal, 'f', 2, 64))
				case "start":
					exists := db.UserExists(database, m.Author.ID)
					if exists {
						s.ChannelMessageSend(m.ChannelID, "You have already registered")
					} else {
						err := db.NewUser(database, m.Author.ID)
						if !err {
							s.ChannelMessageSend(m.ChannelID, "You have been registered")
						} else {
							s.ChannelMessageSend(m.ChannelID, "Failed to register user")
						}
					}

				case "buy":
					if len(args) >= 5 {
						if db.StockExists(database, args[3]) {
							stock_name := args[3]
							var total_investment float64
							var money_symbol string = "$"
							first_char := args[4][0]
							if string(first_char) == money_symbol {
								value, err := strconv.ParseFloat(args[4][1:], 64)
								if err != nil {
									s.ChannelMessageSend(m.ChannelID, "Please input in the correct format.")
								} else {
									total_investment = value
								}
							} else {
								stock_value := db.GetStockPrice(database, stock_name)
								value, err := strconv.ParseFloat(args[4], 64)
								if err != nil {
									s.ChannelMessageSend(m.ChannelID, "Please input in the correct format.")
								} else {
									total_investment = value * stock_value
								}
							}
							investment := total_investment * (1 - tax)

							investment_string := strconv.FormatFloat(investment, 'f', 2, 64)

							msg := discordgo.MessageSend{
								Content: "Are you sure you'd like to invest $" + investment_string + " into " + args[3] + " stocks?",
								Components: []discordgo.MessageComponent{
									yn_button_row,
								},
								Reference: m.Reference(),
							}
							reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

							if err != nil {
								log.Print(err)
							}

							decided := false
							buttonHandler := sess.AddHandler(func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
								if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID && !decided {
									custom_id := button_i.MessageComponentData().CustomID
									content := "Error occured."
									if custom_id == "yes_button" {
										_, content = db.StockTransaction(database, m.Author.ID, args[3], investment_string, 1, tax)
									} else if custom_id == "no_button" {
										content = "Cancelled stock purchase."
									}
									resp := discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseUpdateMessage,
										Data: &discordgo.InteractionResponseData{
											Content: content,
										},
									}
									err := button_s.InteractionRespond(button_i.Interaction, &resp)

									if err != nil {
										log.Println(err)
									}
									decided = true
								}
							})

							// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
							time.AfterFunc(60*time.Second, func() {
								buttonHandler()
							})
						} else {
							s.ChannelMessageSend(m.ChannelID, `"`+args[3]+`" stock does not exist.`)
						}
					}
				case "sell":
					if len(args) >= 5 {
						if db.StockExists(database, args[3]) {
							stock_name := args[3]
							investment := -1.0
							var money_symbol string = "$"
							first_char := args[4][0]
							if string(first_char) == money_symbol {
								value, err := strconv.ParseFloat(args[4][1:], 64)
								if err != nil {
									s.ChannelMessageSend(m.ChannelID, "Please input in the correct format.")
								} else {
									investment = value
								}
							} else if args[4] == "all" {
								stock_value := db.GetStockPrice(database, stock_name)
								total_shares := db.GetUserShares(database, m.Author.ID, stock_name)
								investment = db.PreciseMult(stock_value, total_shares) / (1 - tax)
							} else {
								stock_value := db.GetStockPrice(database, stock_name)
								value, err := strconv.ParseFloat(args[4], 64)
								if err != nil {
									s.ChannelMessageSend(m.ChannelID, "Please input in the correct format.")
								} else {
									investment = db.PreciseMult(value, stock_value) / (1 - tax)
								}
							}
							if investment != -1.0 {
								investment = investment * (1 - tax)
								investment_string := strconv.FormatFloat(investment, 'f', 2, 64)

								msg := discordgo.MessageSend{
									Content: "Are you sure you'd like to sell $" + investment_string + " worth " + args[3] + " stocks?",
									Components: []discordgo.MessageComponent{
										yn_button_row,
									},
									Reference: m.Reference(),
								}
								reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

								if err != nil {
									log.Print(err)
								}

								decided := false
								buttonHandler := sess.AddHandler(func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
									if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID && !decided {
										custom_id := button_i.MessageComponentData().CustomID
										content := "Error occured."
										if custom_id == "yes_button" {
											_, content = db.StockTransaction(database, m.Author.ID, args[3], investment_string, -1, tax)
										} else if custom_id == "no_button" {
											content = "Cancelled stock sale."
										}
										resp := discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseUpdateMessage,
											Data: &discordgo.InteractionResponseData{
												Content: content,
											},
										}
										err := button_s.InteractionRespond(button_i.Interaction, &resp)

										if err != nil {
											log.Println(err)
										}
										decided = true
									}
								})

								// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
								time.AfterFunc(60*time.Second, func() {
									buttonHandler()
								})
							}
						} else {
							s.ChannelMessageSend(m.ChannelID, `"`+args[3]+`" stock does not exist.`)
						}
					}
					/*
						case "buy":
							if db.StockExists(database, args[3]) {

								msg := discordgo.MessageSend{
									Content: "Are you sure you'd like to invest $" + args[4] + " into " + args[3] + " stocks?",
									Components: []discordgo.MessageComponent{
										yn_button_row,
									},
									Reference: m.Reference(),
								}
								reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

								if err != nil {
									log.Print(err)
								}

								decided := false
								buttonHandler := sess.AddHandler(func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
									if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID && !decided {
										custom_id := button_i.MessageComponentData().CustomID
										content := "Error occured."
										if custom_id == "yes_button" {
											_, content = db.StockTransaction(database, m.Author.ID, args[3], args[4], 1)
										} else if custom_id == "no_button" {
											content = "Cancelled stock purchase."
										}
										resp := discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseUpdateMessage,
											Data: &discordgo.InteractionResponseData{
												Content: content,
											},
										}
										err := button_s.InteractionRespond(button_i.Interaction, &resp)

										if err != nil {
											log.Println(err)
										}
										decided = true
									}
								})

								// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
								time.AfterFunc(60*time.Second, func() {
									buttonHandler()
								})

							}
						case "sell":
							if db.StockExists(database, args[3]) {
								msg := discordgo.MessageSend{
									Content: "Are you sure you'd like to sell $" + args[4] + " worth of " + args[3] + " stocks?",
									Components: []discordgo.MessageComponent{
										yn_button_row,
									},
									Reference: m.Reference(),
								}
								reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

								if err != nil {
									log.Print(err)
								}

								decided := false
								buttonHandler := sess.AddHandler(func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
									if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID && !decided {
										custom_id := button_i.MessageComponentData().CustomID
										content := "error occured."

										if custom_id == "yes_button" {
											_, content = db.StockTransaction(database, m.Author.ID, args[3], args[4], -1)

										} else if custom_id == "no_button" {
											content = "cancelled stock sale."
										}

										resp := discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseUpdateMessage,
											Data: &discordgo.InteractionResponseData{
												Content: content,
											},
										}
										err := button_s.InteractionRespond(button_i.Interaction, &resp)

										if err != nil {
											log.Println(err)
										}
										decided = true
									}
								})

								// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
								time.AfterFunc(60*time.Second, func() {
									buttonHandler()
								})

							}
					*/
				case "status":
					if len(args) == 4 {
						response := db.ReturnStock(database, args[3])
						s.ChannelMessageSend(m.ChannelID, response)
					} else if len(args) == 3 {
						response := db.ReturnStock(database, "*")
						s.ChannelMessageSend(m.ChannelID, response)
					}
				case "list":
					response := db.UserList(database, m.Author.ID)
					s.ChannelMessageSend(m.ChannelID, response)
				case "networth":
					response := db.GetNetWorth(database, m.Author.ID)
					s.ChannelMessageSend(m.ChannelID, response)
				case "help":
					response := helpMessage
					if len(args) > 3 {
						for _, e := range helpSlice {
							if e.Command == args[3] {
								response = e.Explanation
							}
						}
					}
					s.ChannelMessageSend(m.ChannelID, "``"+response+"``")
				}

			}
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()

	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("The bot is online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

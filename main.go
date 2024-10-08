package main

// imports
import (
	"encoding/json"
	"fmt"
  "math"
	db "agendamarket/database"
  "agendamarket/formatcheck"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
  _ "github.com/go-sql-driver/mysql"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// prefix for each command
const prefix string = "!jdl"


// help struct for each command
type Help struct {
	Command     string
	Format      string
	Explanation string
}

func main() { 
  // loading the database
	err := godotenv.Load()
	if err != nil {
		log.Fatal("couldn't load .env file")
	}


  // initiating the bot using the token
	sess, err := discordgo.New("Bot " + os.Getenv(("TOKEN")))
	if err != nil{
		log.Fatal(err)
	}
  // initialising database
	database := db.Init()
	helpFileBytes, err := os.ReadFile("./src/help.json")
	if err != nil {
		log.Fatal(err)
	}

  // getting the information for every command from the JSON file.
	var helpSlice []Help

	err = json.Unmarshal(helpFileBytes, &helpSlice)

	if err != nil {
		log.Fatal(err)
	}

  // creating a list filled with each embed field.
  var help_fields []*discordgo.MessageEmbedField


  // creating the help fields.
  for _, e := range helpSlice {
    field := &discordgo.MessageEmbedField{
      Name:  e.Command,
      Value: e.Format,
    }
    help_fields = append(help_fields, field)
  }

  // help embed
  help_embed := &discordgo.MessageEmbed{}
  help_embed.Fields = help_fields	
  help_embed.Title = "Command Master List"

  
  // initialising every button type.
  // yes button
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
  leftmost_button := discordgo.Button {
    Label: "<<",
    CustomID: "leftmost",
  }
	left_button := discordgo.Button{
		Label:    "<",
		CustomID: "left",
	}
	right_button := discordgo.Button{
		Label:    ">",
		CustomID: "right",
	}
  rightmost_button := discordgo.Button {
    Label: ">>",
    CustomID: "rightmost",
  }
	lr_button_row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{leftmost_button, left_button, right_button, rightmost_button},
	}

	tax, _ := strconv.ParseFloat(os.Getenv("TAX"), 64)
  // creating an event handler for messages received.
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
    // ignoring the message if it was sent by the bot.
		if m.Author.ID == s.State.User.ID {
			return
		}
    // printing the content of the message to the console.
		fmt.Println(m.Content)
    // making a stock list
		full_stock_list := db.ReturnStock(database, "*")
		sort.Slice(full_stock_list, func(i, j int) bool {
			item1, _ := strconv.ParseFloat(full_stock_list[i][1], 64)
			item2, _ := strconv.ParseFloat(full_stock_list[j][1], 64)
			return item1 > item2
		})

		args := strings.Split(m.Content, " ")

		
    for i, arg := range args {
      args[i] = strings.ToLower(arg)
    }
    // ignoring the content of the message if the command wasn't used.
    if args[0] != prefix {
			return
		}
    
		if len(args) > 1 {
      // dealing with every command type.
			switch args[1] {
			case "new":
				if m.Author.ID == os.Getenv("ADMIN") && len(args) > 3 {
					price, err := strconv.ParseFloat(args[3], 64)
					if err == nil {
						cond := db.NewStock(database, args[2], price)
						if cond {
              response := "Add new stock "+args[2]+" with a base price of $"+args[3]
              msg := discordgo.MessageSend{
                Content: response,
                Reference: m.Reference(),
              }
						  _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
              
              if err != nil {
                log.Print(err)
              }

						}
					}
				}
			case "remove":
				if m.Author.ID == os.Getenv("ADMIN") && len(args) > 2 {
					if db.RemoveStock(database, args[2]) {
            response := "Removed stock "+args[2]
            msg := discordgo.MessageSend{
              Content: response,
              Reference: m.Reference(),
            }
            _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
            if err != nil {
              log.Print(err)
            }
					}
				}
			case "balance":
				bal := db.BalCheck(database, m.Author.ID)
        if bal >= 0 {
          response := "Your current balance is $"+strconv.FormatFloat(bal, 'f', 2, 64)
          msg := discordgo.MessageSend{
            Content: response,
            Reference: m.Reference(),
          }
          _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
          if err != nil {
            log.Print(err)
          }
        } else {
          response := "Please register as a new user using the command !jdl register"
          msg := discordgo.MessageSend{
            Content: response,
            Reference: m.Reference(),
          }
          _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
          if err != nil {
            log.Print(err)
          }
        }
			case "register":
				exists := db.UserExists(database, m.Author.ID)
				if exists {
					response := "You have already registered"
          msg := discordgo.MessageSend{
            Content: response,
            Reference: m.Reference(),
          }
          _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
          if err != nil {
            log.Print(err)
          }
				} else {
					err := db.NewUser(database, m.Author.ID, 1000)
					if !err {
						response := "You have been registered"
            msg := discordgo.MessageSend{
              Content: response,
              Reference: m.Reference(),
            }
            _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
            if err != nil {
              log.Print(err)
            }
          } else {
            response := "Failed to register user"
            msg := discordgo.MessageSend{
              Content: response,
              Reference: m.Reference(),
            }
            _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
            
            if err != nil {
              log.Print(err)
            }
					}
				}

			case "buy":
        // checking of the format is valid
        response, valid := formatcheck.CheckFormat("buy", args, database)
      
        // if the format is valid, the transaction is attempted.
				if valid {
            // extracting the stock name, the value given and the investment.
						var total_investment float64 = -1.0
						var money_symbol string = "$"
						first_char := args[3][0]

            // if the investment is specified in dollars, the value is extracted from after the first character of the fourth argument.
						if string(first_char) == money_symbol {
							value, err := strconv.ParseFloat(args[3][1:], 64)
							if err != nil {
                log.Println(err)
							} else {
                total_investment = value
              }

						} else {
              value, err := strconv.ParseFloat(args[3], 64)
              if err != nil {
                log.Println(err)
              } else {
                  total_investment = value
              }
						}

            // if the total investment is something valid (which will not be -1), the transaction goes forward.
            if total_investment != -1.0 {
						investment := total_investment * (1 - tax)

						investment_aftertax_string := strconv.FormatFloat(investment, 'f', 2, 64)
						investment_string := strconv.FormatFloat(
							total_investment,
							'f',
							2,
							64,
						)

						msg := discordgo.MessageSend{
							Content: "Are you sure you'd like to invest $" + investment_string + " into buying $" + investment_aftertax_string + " worth of " + args[2] + " stocks?",
							Components: []discordgo.MessageComponent{
								yn_button_row,
							},
							Reference: m.Reference(),
						}
						reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

						if err != nil {
							log.Print(err)
						}

						buttonHandler := sess.AddHandler(
							func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
                log.Print(m.Author.ID)
								if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID {
									custom_id := button_i.MessageComponentData().CustomID
									content := "Error occured."
									if custom_id == "yes_button" {
										_, content = db.StockTransaction(
											database,
											m.Author.ID,
											args[2],
											investment_aftertax_string,
											1,
											tax,
										)
									} else if custom_id == "no_button" {
										content = "Cancelled stock purchase."
									}
                  // creating the response

									resp := discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseUpdateMessage,
										Data: &discordgo.InteractionResponseData{
											Content: content,
										},
									}
                  // responding while also checking for a potential error.
									err := button_s.InteractionRespond(button_i.Interaction, &resp)

									if err != nil {
										log.Println(err)
									}
								}
							},
						)

						// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
						time.AfterFunc(60*time.Second, func() {
							buttonHandler()
						})
            }
				} else {
          msg := discordgo.MessageSend{
            Content: response,
            Reference: m.Reference(),
          }
          _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

          if err != nil {
            log.Print(err)
          }
        }
			case "sell":
        // checking of the format is valid
        response, valid := formatcheck.CheckFormat("sell", args, database)
      
        // if the format is valid, the transaction is attempted.
				if valid {
            // extracting the stock name, the value given and the investment.
						stock_name := args[2]
						var total_investment float64 = -1.0
						var money_symbol string = "$"
						first_char := args[3][0]

            // if the investment is specified in dollars, the value is extracted from after the first character of the fourth argument.
						if string(first_char) == money_symbol {
							value, err := strconv.ParseFloat(args[3][1:], 64)
							if err != nil {
                log.Println(err)
							} else {
                total_investment = value
              }

						} else if args[3] != "all" {
              // otherwise, the stock value is extracted by multiplying the stock value
							stock_value := db.GetStockPrice(database, stock_name)
							value, err := strconv.ParseFloat(args[3], 64)
							if err != nil {
						    log.Println(err)	
							} else {
                total_investment = value * stock_value
							}

						} else {
              stock_value := db.GetStockPrice(database, stock_name)
              user_stocks := db.GetUserShares(database, m.Author.ID, stock_name)
              total_investment = stock_value * user_stocks
            }
            // if the investment is valid (not == -1) then it is converted into a string 
						if total_investment != -1.0 {
							investment_string := strconv.FormatFloat(total_investment, 'f', 2, 64)
							investment_aftertax := strconv.FormatFloat(
								total_investment*(1-tax),
								'f',
								2,
								64,
							)

							msg := discordgo.MessageSend{
								Content: "Are you sure you'd like to $" + investment_string + " worth " + args[2] + " stocks? $" + investment_aftertax + " will be added to your balance",
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


              // the event handler for the message that was just sent, waits for the response of the user.
							buttonHandler := sess.AddHandler(
                // function initialisation
								func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
                  // checks to see if the message ID and the button ID are of the message that was just sent.
									if button_i.Message.ID == reply.ID && m.Author.ID == button_i.Member.User.ID && !decided {
                    // getting the ID of the button which was pressed. 
										custom_id := button_i.MessageComponentData().CustomID
                    // Initialising content as "error occured", although it will likely change.
										content := "Error occured."
                    // if custom ID == yes, the user wants to buy, so the transaction method is called from db..
										if custom_id == "yes_button" {
											_, content = db.StockTransaction(
												database,
												m.Author.ID,
												args[2],
												investment_string,
												-1,
												tax,
											)
										} else if custom_id == "no_button" {
                      // cancelling the stock sale if the no button is pressed
											content = "Cancelled stock sale."
										}
                    // generating the response variable. 
										resp := discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseUpdateMessage,
											Data: &discordgo.InteractionResponseData{
												Content: content,
											},
										}
                    // responding to the interaction.
										err := button_s.InteractionRespond(
											button_i.Interaction,
											&resp,
										)

										if err != nil {
											log.Println(err)
										}

										decided = true
									}
								},
							)

							// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
							time.AfterFunc(60*time.Second, func() {
								buttonHandler()
							})
						}
					} else {
            // responding if the stock doesn't exist.
						msg := discordgo.MessageSend{
              Content: response,
              Reference: m.Reference(),
            }
            _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

            if err != nil {
              log.Print(err)
            }
          }
          
				
      case "list":
        message := db.UserList(database, m.Author.ID)
        
        
        if message == "" {
          message = "You do not currently own any stocks."
        }
        msg := discordgo.MessageSend{
          Content: message,
          Reference: m.Reference(),
        }
        _, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)

        if err != nil {
          log.Print(err)
        }        
			case "shares":
				var stock_list [][]string

				conditionCheck := func() bool {
					if len(args) == 2 {
						return true
					}
					if !db.StockExists(database, args[2]) {
						return true
					}
					return false
				}

				if conditionCheck() {

					stock_list = full_stock_list

					embed := &discordgo.MessageEmbed{}
					embed.Title = "Stock Prices per share"
					embed.Description = "Use the left and right arrows to display the next set of stocks."
					current_index := 0
					fieldsperembed := 10
					stock_length := len(stock_list)
					max_index := int(math.Ceil(float64(stock_length) / float64(fieldsperembed)))


          // creating a list filled with each embed field.
					var all_fields []*discordgo.MessageEmbedField
					for _, e := range stock_list {
            if e[1] != "" {
              field := &discordgo.MessageEmbedField{
                Name:  e[0],
                Value: "$" + e[1] + " per share",
              }
              all_fields = append(all_fields, field)
            }
					}

          // creating the message object consisting of the embeds and the buttons.
					msg := discordgo.MessageSend{
						Components: []discordgo.MessageComponent{
							lr_button_row,
						},
            Reference: m.Reference(),
					}
          // making a list of each page.
					lower_index := current_index * fieldsperembed
					upper_index := (current_index + 1) * fieldsperembed
					if upper_index >= stock_length {
						upper_index = stock_length
					}
					embed.Fields = all_fields[lower_index:upper_index]
					msg.Embeds = append(msg.Embeds, embed)
          // sending the message using "SendComplex" in order to have the option to edit the message everytime someone changes the page.
					reply, err := s.ChannelMessageSendComplex(m.ChannelID, &msg)
					if err != nil {
						log.Print(err)
					}

          // button handler for the right left buttons.
					buttonHandler := sess.AddHandler(
						func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
              // checking if the button pressed was both from the right message and the right user.
							if button_i.Message.ID == reply.ID &&
								m.Author.ID == button_i.Member.User.ID {
								CustomID := button_i.MessageComponentData().CustomID
                // if the left arrow was pressed, go the the previous page.
								if CustomID == "left" {
									current_index -= 1
									if current_index < 0 {
										current_index = max_index - 1
									}
                // if the right arrow was pressed, go to the next page.
								} else if CustomID == "right" {
									current_index = (current_index + 1) % max_index
								} else if CustomID == "leftmost" {
                  current_index = 0
                } else if CustomID == "rightmost" {
                  current_index = max_index - 1
                }
								lower_index := current_index * fieldsperembed
								upper_index := (current_index + 1) * fieldsperembed
								if upper_index >= stock_length {
									upper_index = stock_length
								}
                // editing the embed when the button was pressed.
								embed.Fields = all_fields[lower_index:upper_index]
								resp := discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseUpdateMessage,
									Data: &discordgo.InteractionResponseData{
										Components: []discordgo.MessageComponent{
											lr_button_row,
										},
										Embeds: []*discordgo.MessageEmbed{
											embed,
										},
									},
								}
								button_s.InteractionRespond(button_i.Interaction, &resp)
							}
						},
					)

					// Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
					time.AfterFunc(60*time.Second, func() {
						buttonHandler()
					})
          // if a specific stock was searched for, only the stock that was requested is returned to the user.
				} else if db.StockExists(database, args[2]) {
					price := db.GetStockPrice(database, args[2])
          response := args[1] + " stock is currently worth $" + strconv.FormatFloat(price, 'f', 2, 64)  + " per share."
          msg := discordgo.MessageSend{
            Content: response,
            Reference: m.Reference(),
          }
          s.ChannelMessageSendComplex(m.ChannelID, &msg)
				}

			case "networth":
        // getting the networth and sending it to the user.
				response := db.GetNetWorth(database, m.Author.ID)
        msg := discordgo.MessageSend{
          Content: response,
          Reference: m.Reference(),
        }
				s.ChannelMessageSendComplex(m.ChannelID, &msg)
			case "help":
        // generating a help message and sending it.
        msg := discordgo.MessageSend{
          Embeds: []*discordgo.MessageEmbed{
            help_embed,
          },
          Reference: m.Reference(),
        }
				s.ChannelMessageSendComplex(m.ChannelID, &msg)
			}
		} else {
      msg := discordgo.MessageSend{
        Embeds: []*discordgo.MessageEmbed{
          help_embed,
        },
        Reference: m.Reference(),
      }
      s.ChannelMessageSendComplex(m.ChannelID, &msg)
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

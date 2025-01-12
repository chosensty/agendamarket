package command_handler

import (
	"github.com/bwmarrin/discordgo"
)

var (
	BuyCommand = Command{
		Name: "Buy",
		Command: discordgo.ApplicationCommand{
			Name:        "Buy",
			Description: "Buy a stock",
      Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Please enter the name of the stock that you'd like to purchase",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "investment",
					Description: "Please enter the amount that you'd like to invest",
          MinValue: 0,
					Required:    true,
				},
			},
		},
		Execute: buyCommand,
	}
)

func buyCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {// extracting the data from the interaction response
  opts := i.applicationcommanddata().options

  options := make(map[string]*discordgo.applicationcommandinteractiondataoption, len(opts))
  for _, opt := range opts {
    options[opt.name] = opt
  }
  stock_name = options["name"]
  investment = options["investment"]


  // calculating the investment after tax
  tax_rate = os.Getenv("TAX")
  i_aftertax = investment * (1 - tax_rate)


  response := ""

  if !db.StockExists(database, stock_name) {
    response = "Could not find \"" + stock_name + "\" stock."
  } else {
    user_bal = db.UserBalance(i.Member.User.ID)
    if user_bal < investment {
      response = "I don't think you have the facilities for that big man."
    } else {
      message := "Are you sure you'd like to invest $" + investment + " into " + stock_name + " stocks?"

      // getting user confirmation
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Components: []discordgo.MessageComponent{
            yn_button_row,
          },
          Content: message,
        },
      })

      decided := false
      // the event handler for the message that was just sent, waits for the response of the user.
      buttonHandler := sess.AddHandler(
        func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
          if button_i.Message.ID == reply.ID && i.Member.User.ID == button_i.Member.User.ID && !decided {

            custom_id := button_i.MessageComponentData().CustomID
            content := ""

            if custom_id == "yes_button" {
              // completing the transaction
              content = db.StockTransaction()

            } else if custom_id == "no_button" {
              // cancelling the stock sale if the no button is pressed
              content = "Cancelled stock sale."
            }
            // responding to the user.
            s.InteractionResponseEdit{
              Type: discordgo.InteractionResponseUpdateMessage,
              Data: &discordgo.InteractionResponseData{
                Content: content,
              },
            }
            decided = true
          }
        },
      )

      // Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
      time.AfterFunc(60*time.Second, func() {
        buttonHandler()

        if decided == false {
          
          message := "Request timed out"
          s.InteractionResponseEdit(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
              Components: []discordgo.MessageComponent{
                yn_button_row,
              },
              Content: message,
            },
          })
        }
      })
      
      
    }
    if response != "" {
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Components: []discordgo.MessageComponent{
            yn_button_row,
          },
          Content: message,
        },
      })
    }

    
  }

}

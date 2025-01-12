package command_handler

import (
	"github.com/bwmarrin/discordgo"
  "math"
)

var (
	PricesCommand = Command{
		Name: "Prices",
		Command: discordgo.ApplicationCommand{
			Name:        "Prices",
			Description: "List all of the available stocks and their prices",
      Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "Stock name",
					Description: "Find the price of a specific stock",
					Required:    false,
				},
			},
		},
		Execute: pricesCommand,
	}
)

func pricesCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

  // extracting the data from the interaction response
  opts := i.applicationcommanddata().options
  options := make(map[string]*discordgo.applicationcommandinteractiondataoption, len(opts))
  for _, opt := range opts {
    options[opt.name] = opt
  }

  stock_name = options["stock name"]


  if stock_name == "" {

    response := ""
    if !db.StockExists(database, stock_name) {
      response = "Could not find stock " + stock_name
      
    } else {
      price := db.GetStockPrice(database, stock_name)
      response :=  stock_name + " is currently worth $" + strconv.FormatFloat(price, 'f', 2, 64)  + " per share."
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
      Type: discordgo.InteractionResponseChannelMessageWithSource,
      Data: &discordgo.InteractionResponseData{
        Content: response,
      },
    })


  } else {
    embed := &discordgo.MessageEmbed{}
 // making a stock list for each item
		stock_list := db.ReturnStock(database, "*")
		sort.Slice(stock_list, func(i, j int) bool {
			item1, _ := strconv.ParseFloat(stock_list[i][1], 64)
			item2, _ := strconv.ParseFloat(stock_list[j][1], 64)
			return item1 > item2
		})

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
            Value: "$" + e[1],
          }
          all_fields = append(all_fields, field)
        }
      }

      // making a list of each page.
      lower_index := current_index * fieldsperembed
      upper_index := (current_index + 1) * fieldsperembed
      if upper_index >= stock_length {
        upper_index = stock_length
      }

      embed.Fields = all_fields[lower_index:upper_index]


      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Embeds: []*discordgo.MessageEmbed{
            embed,
          },
        },
      }) 

      // button handler for the right left buttons.
      buttonHandler := sess.AddHandler(
        func(button_s *discordgo.Session, button_i *discordgo.InteractionCreate) {
          // checking if the button pressed was both from the right message and the right user.
          if button_i.Message.ID == reply.ID &&
            i.Member.User.ID == button_i.Member.User.ID {
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

            // updating the original message.
            resp := s.InteractionResponseEdit{
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

          }
        },
      )

      // Set a timer to stop listening after a specific amount of time (e.g., 30 seconds)
      time.AfterFunc(60*time.Second, func() {
        buttonHandler()
      })
				}
}

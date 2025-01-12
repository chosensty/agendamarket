package command_handler

import (
	"github.com/bwmarrin/discordgo"
)

var (
	SellCommand = Command{
		Name: "Sell",
		Command: discordgo.ApplicationCommand{
			Name:        "Sell",
			Description: "Sell a stock",
      Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Please enter the name of the stock that you'd like to sell",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "sale",
					Description: "Please enter the amount of shares that you'd like to sell",
          MinValue: 0,
					Required:    true,
				},
			},
		},
		Execute: sellCommand,
	}
)

func sellCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
// extracting the data from the interaction response
  opts := i.applicationcommanddata().options
  options := make(map[string]*discordgo.applicationcommandinteractiondataoption, len(opts))
  for _, opt := range opts {
    options[opt.name] = opt
  }

}

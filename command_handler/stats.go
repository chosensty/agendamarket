package command_handler

import (
	"github.com/bwmarrin/discordgo"
)

var (
	StatsCommand = Command{
		Name: "Stats",
		Command: discordgo.ApplicationCommand{
			Name:        "Stats",
			Description: "View your balance and every stock that you own",

		},
		Execute: statsCommand,
	}
)

func statsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
  reponse := ""

  bal := db.BalCheck(database, m.Author.ID)

  if bal >= 0 {
    response = "Your current balance is $"+strconv.FormatFloat(bal, 'f', 2, 64)

    // getting the user stock list
    message := db.UserList(database, m.Author.ID)

    if message = "" {
      message = "You do not currently own any stocks."
    }

    response = response + "\n" + message
            
    
  } else {
    response = "Please register as a new user using the command /register"
  }

  s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: response,
    },
  }) 
}

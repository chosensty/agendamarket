package command_handler

import (
	"github.com/bwmarrin/discordgo"
)

var (
	RegisterCommand = Command{
		Name: "Register",
		Command: discordgo.ApplicationCommand{
			Name:        "Register",
			Description: "Register as a new user",
		},
		Execute: registerCommand,
	}
)

func registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
  exists := db.UserExists(database, i.Member.User.ID)
  
  
  response := ""

  // if the user already exists, return a message.
  if exists {
    response := "You have already registered"
  } else {
    err := db.NewUser(database, i.Member.User.ID, 1000)

    if !err {
      response := "You have been registered"
    } else {
      response := "Failed to register user"
    }
  }


  s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: response,
    },
  })
}

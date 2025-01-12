package command_handler

import (
	"os"
	"strings"
	"github.com/bwmarrin/discordgo"
)

var (
  // array of all the different commands
	commands = []Command{
		BuyCommand,
		SellCommand,
		StatsCommand,
		StocksCommand,
    RegisterCommand,
	}
)

type Command struct {
	Name    string
	Command discordgo.ApplicationCommand
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func StartCommandHandler(s *discordgo.Session) {
	var appID string

  appID = os.Getenv("APP_ID")

	for _, command := range commands {
		cmd, err := s.ApplicationCommandCreate(appID, "", &command.Command)
		if err != nil {
			panic(err)
		}
		fmt.Println("Registered command /" + cmd.Name)
	}
	fmt.Println("Commands all successfully registered")
}

// creating a command handler
func InteractionCreateListener(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		cmdData := i.ApplicationCommandData()

		for _, command := range commands {
      // checking which command was ran and executing that command
			if cmdData.Name == command.Name {
				command.Execute(s, i)
			}
		}
  }
}

package event_listener

import (
	"agendamarket/command_handler"
  "fmt"
	"github.com/bwmarrin/discordgo"
)

func StartEventListeners(s *discordgo.Session) {
	s.AddHandler(ready)
	s.AddHandler(commands.InteractionCreateListener)

  fmt.Println("Command handler is ready")
}

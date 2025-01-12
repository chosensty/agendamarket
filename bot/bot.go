package bot

import (
  "agendamarket/event_listener"
  "agendamarket/command_handler"
	"os"
  "fmt"
  "log"
	"os/signal"
	"syscall"
	"github.com/bwmarrin/discordgo"
)

func StartBot() {
	var token string
  token = os.Getenv("TOKEN")
	

	s, err := discordgo.New("Bot " + token)
	if err != nil {
    log.Fatal("Error starting the bot: " + err.Error())
	}

	e_handler.StartEventListeners(s)
	c_handler.StartCommandHandler(s)

	if err != nil {
    log.Fatal("Error occured while initiating handlers: " + err.Error())
	}

	err = s.Open()
	if err != nil {
		log.Fatal("Error starting discord session: " + err.Error())
	}

	notifs.System("Bot is now online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	fmt.Println("Bot has been stopped")
	s.Close()
}

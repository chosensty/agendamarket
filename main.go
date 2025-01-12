package main

import (
  "agendamarket/bot"
  "github.com/joho/godotenv"
)

// loading environmental variables and initiating bot start up
func main() {
  err := godotenv.Load()

  if err != nil {
    panic(err)
  }

  go globals.UpdateCaches()

  bot.StartBot()
}

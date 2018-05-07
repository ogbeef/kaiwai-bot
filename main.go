package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Member variables
var (
	Token string
)

// @package
// package initialization
func init() {
	//Parse command line arguments.
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

// @package
// package entory point
func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if !checkError(err, "error creating Discord session,") {
		return
	}

	// Set eventParser to session.
	dg.AddHandler(eventParser)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if !checkError(err, "error opening connection,") {
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// @fn
// Parse Discord events.
// @param
func eventParser(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if message.Author.ID == session.State.User.ID {
		return
	}

	// PingPong Command.
	if checkRegexp("^(p|P)ing", message.Content) {
		pingPongEvent(session, message)
	}
}

// @fn
// Check error.
// @param err : Error handler.
// @param message : Error message which is dumped to log.
// @return : Return false if error is happend.
func checkError(err error, message string) bool {
	//error check
	if err != nil {
		fmt.Println(message, err)
		return false
	}
	return true
}

// @fn
// Check regexp.
// @param reg : Regular Expression.
// @param str : Target string.
// @return : Return true if the regexp is match.
func checkRegexp(reg string, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}

// @fn
// Ping Pong Event.
// @param session : Discord session.
// @param message : Received message.
func pingPongEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	session.ChannelMessageSend(message.ChannelID, "Pong!")
}

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

// Constants
const AnicoBinUrl = "http://anicobin.ldblog.jp/"

// Struct
// Anico Web Page
type AnicoWebPage struct {
	Title     string
	Link      string
	Thumbnail string
}

// Member variables
var (
	Token string
)

// @package
// package initialization
func init() {
	// Parse command line arguments.
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	// Set random seed.
	rand.Seed(time.Now().UnixNano())
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
	// Anime Capture Command.
	if checkRegexp("^(a|A)nime", message.Content) {
		animeCaptureEvent(session, message)
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
	fmt.Println("event PINGPONG from ", message.Author.Username)
	session.ChannelMessageSend(message.ChannelID, "Pong!")
}

// @fn
// Anime Capture Event.
// @param session : Discord session.
// @param message : Received message.
func animeCaptureEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	fmt.Println("event ANIME_CAPTURE from ", message.Author.Username)
	session.ChannelMessageSend(message.ChannelID, "Anime!")

	var animeList []AnicoWebPage

	//Load 5 pages
	for pageIndex := 0; pageIndex < 5; pageIndex++ {
		// Create target url.
		targetUrl := AnicoBinUrl + "/?p=" + fmt.Sprint(pageIndex+1)
		// Request anico-bin.
		response, err := http.Get(targetUrl)
		if !checkError(err, "error connecting anico-bin,") {
			session.ChannelMessageSend(message.ChannelID, "Access Error!")
			return
		}
		defer response.Body.Close()
		// Check status code.
		// If return code is not 200, failed to access.
		if response.StatusCode != 200 {
			fmt.Println("error, status code error: %d %s", response.StatusCode, response.Status)
			session.ChannelMessageSend(message.ChannelID, "Access Error!")
			return
		}
		// Load the HTML document.
		anico, err := goquery.NewDocumentFromReader(response.Body)
		if !checkError(err, "error getting document body,") {
			session.ChannelMessageSend(message.ChannelID, "Access Error!")
			return
		}

		// Get anime list.
		anico.Find(".ArticleFirstImageThumbnail").Each(func(i int, s *goquery.Selection) {
			// Parse
			// ToDo : Need refactor
			link, _ := s.Find("a").Attr("href")
			thum, _ := s.Find("img").Attr("src")
			title, _ := s.Find("img").Attr("alt")
			imageUrl := strings.Split(thum, "http://")
			if len(imageUrl) >= 3 {
				var page AnicoWebPage
				page.Link = link
				page.Title = title
				page.Thumbnail = "http://" + imageUrl[2]
				if !checkRegexp("(2018|2017)", page.Title) {
					animeList = append(animeList, page)
				}

			}
		})
	}

	// Dump parsing anime list.
	for index, anime := range animeList {
		fmt.Println(index, " title:", anime.Title, " link:", anime.Link, " Thumbnail:", anime.Thumbnail)
	}

	// Select anime!
	dice := rand.Intn(len(animeList) - 1)
	fmt.Println("selected anime = ", dice, " title:", animeList[dice].Title, " link:", animeList[dice].Link, " Thumbnail:", animeList[dice].Thumbnail)

	// Send message with thumbnail url
	ret := "Title:" + animeList[dice].Title + "\n" + animeList[dice].Thumbnail
	session.ChannelMessageSend(message.ChannelID, ret)
}

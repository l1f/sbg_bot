package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/mattn/go-mastodon"
)


const userAgent = "SBG bot nörgel bot by @lou@queer.group"

type flags struct {
  server string
  clientId string
  clientSecret string
  accessToken string
  prevFile string
}

func parseArgs() (*flags, error) {
	var flagSet      = flag.NewFlagSet("sbg-nörgel-bot", flag.ExitOnError)
	var server       = flagSet.String("server", "https://queer.group", "The URL of the Mastodon server")
	var clientID     = flagSet.String("client-id", "", "The client ID for the application")
	var clientSecret = flagSet.String("client-secret", "", "The client secret for the application")
	var accessToken  = flagSet.String("access-token", "", "The access token for the bot account")
  var prevFile     = flagSet.String("prev-toot-file", "", "The file in which the last toot id is stored")

  if err := flagSet.Parse(os.Args[1:]); err != nil {
	  return nil, fmt.Errorf("parsing arguments failed: %w", err)
  }
  
	if *server == "" {
    return nil, errors.New("the -server flag is empty, please provide a value")
	}

	if *clientID == "" {
		return nil, errors.New("the -client-id flag is empty, please provide a value")
	}

	if *clientSecret == "" {
		return nil, errors.New("the -client-secret flag is empty, please provide a value")
	}

	if *accessToken == "" {
		return nil, errors.New("the -access-token flag is empty, please provide a value")
	}

  if *prevFile == "" {
		return nil, errors.New("the -prev-toot-file flag is empty, please provide a value")
  }

 parsedFlags := flags{
    server: *server,
    clientId: *clientID,
    clientSecret: *clientSecret,
    accessToken: *accessToken,
    prevFile: *prevFile,
  }

  return &parsedFlags, nil
}

func main() {
  if err := run(); err != nil {
    log.Fatal(err)
  }
}

func run() error {
  parsedFlags, err := parseArgs()
  if err != nil {
    return err
  }

  client := mastodon.NewClient(&mastodon.Config{
    Server: parsedFlags.server,
    ClientID: parsedFlags.clientId,
    ClientSecret: parsedFlags.clientSecret,
    AccessToken: parsedFlags.accessToken,
  })

  client.UserAgent = userAgent

  t := toot{
    path: parsedFlags.prevFile,
  }
  
  t.load()

  task := func () {
    t.post(*client)
  }

  scheduler := gocron.NewScheduler(time.Local) 
  scheduler.Every(1).Day().At("10:00").Do(task)
  
  log.Println("Start Scheduler")

  scheduler.StartBlocking()

  return nil
}

type toot struct{
  prevToot string
  path string
}

func (t *toot) save(prevToot string) error {
  t.prevToot = prevToot
   	
  data := []byte(prevToot)
  
  return os.WriteFile(t.path, data, 0644)
}

func (t *toot) load() {
  t.prevToot = ""
  data, err := os.ReadFile(t.path)
  if err != nil {
    log.Printf("Can't read %s, error: %s", t.path, err)
    return
  }
  
  t.prevToot = string(data)
}

func (t *toot) post(client mastodon.Client) error {
  var toot = mastodon.Toot{
    Status: buildToot(),
    Language: "de",
    Visibility: "unlisted",
  }

  if t.prevToot != "" {
    toot.InReplyToID = mastodon.ID(t.prevToot)
  }

  log.Println("Post tsg nörgel")

  status, err := client.PostStatus(context.Background(), &toot)

  if err != nil {
    log.Printf("Error post status: %s", err)
  }

  t.save(string(status.ID))

  return nil
}

func calculateDays() int {
  var pause = time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local)
  var today = time.Now()
  
  hoursUntil := pause.Sub(today).Hours()
  
  return int(hoursUntil / 24)
}

const tootTemplate = "Daily Reminder, es sind noch %d Tage bis zur Parlamentarischen Sommerpause. Bis dahin möchten #SPD, #Grüne und #FDP das diskriminierende TSG durch ein Selbstbestimmungsgesetz ersetzen."

func buildToot() string {
  return fmt.Sprintf(tootTemplate, calculateDays())
}

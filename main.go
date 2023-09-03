package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"

	static "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

type chatMsg struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

var csvFile *os.File

func main() {
	r := gin.Default()
	m := melody.New()
	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	csvFilePath := "chat.csv"

	// Prüfe, ob die CSV-Datei existiert
	_, err := os.Stat(csvFilePath)
	if os.IsNotExist(err) {
		if _, err := os.Create(csvFilePath); err != nil {
			log.Error(err)
			return
		}
	} else if err != nil {
		log.Error(err)
		return
	}

	csvFile, err = os.OpenFile(csvFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Error(err)
		return
	}
	defer csvFile.Close()

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleConnect(func(s *melody.Session) {
		csvReader := csv.NewReader(csvFile)
		// records, err := csvReader.Read()
		if err != nil {
			fmt.Println(err)
		}
		var message chatMsg
		for {
			record, err := csvReader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println(err)
				return
			}
			message.Username = record[1]
			message.Content = record[2]
			msg, err := json.Marshal(message)
			if err != nil {
				log.Error(err)
			}
			if err := m.Broadcast(msg); err != nil {
				log.Error(err)
				return
			}
		}
		message.Username = "Server"
		message.Content = "Hello!"
		msg, err := json.Marshal(message)
		if err != nil {
			log.Error(err)
		}
		if err := m.Broadcast(msg); err != nil {
			log.Error(err)
			return
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		csvwriter := csv.NewWriter(csvFile)

		var message chatMsg
		err := json.Unmarshal(msg, &message)
		if err != nil {
			log.Error(err)
			return
		}
		messageData := []string{time.Now().String(), message.Username, message.Content}

		if err := csvwriter.Write(messageData); err != nil {
			log.Error(err)
			return
		}
		csvwriter.Flush()

		m.Broadcast(msg)
	})

	r.Run(":5000")
}

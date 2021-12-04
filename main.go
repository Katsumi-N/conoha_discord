package main

import (
	"conoha/api"
	"conoha/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// discordのトークン
	Token := config.Config.DiscordToken
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)
	dg.AddHandler(Server)
	// dg.AddHandlerOnce(Timer)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
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

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	log.Println("BotName: ", event.User.ID)
	log.Println("BotID: ", event.User.Username)
}

func Server(s *discordgo.Session, m *discordgo.MessageCreate) {
	token := api.GetToken()
	serverStatus, flavorId := api.GetServerStatus(token)
	var flavor string
	if flavorId == config.Config.Flavor1gb {
		flavor = "1gb"
	} else if flavorId == config.Config.Flavor4gb {
		flavor = "4gb"
	} else {
		flavor = flavorId
	}
	if command := m.Content; command == "!server" {
		if serverStatus == "SHUTOFF" {
			s.ChannelMessageSend(m.ChannelID, "サーバーは起動していません")

		} else if serverStatus == "ACTIVE" {
			s.ChannelMessageSend(m.ChannelID, "サーバーは起動中")

		} else {
			fmt.Println(serverStatus)
			s.ChannelMessageSend(m.ChannelID, "再起動中　時間をおいてもう一度やり直してください")
		}
		flavor_message := fmt.Sprintf("メモリタイプ:%s", flavor)
		s.ChannelMessageSend(m.ChannelID, flavor_message)

	} else if command == "!start" {
		if serverStatus == "SHUTOFF" {
			// サーバーのflavorを1gb->4gbに変更
			for {
				err := api.ChangeServerFlavor(token, flavor, "4gb")
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
				}
				statusCheck, flavorChanged := api.GetServerStatus(token)
				log.Println("現在の状態:", statusCheck)
				if statusCheck == "VERIFY_RESIZE" {
					resizeStatusCode, _ := api.ConfirmResize(token)
					for resizeStatusCode != 204 {
						time.Sleep(10 * time.Second) // 10秒待ってから再リクエスト
						resizeStatusCode, _ = api.ConfirmResize(token)
					}

					break
				} else if flavorChanged == config.Config.Flavor4gb {
					s.ChannelMessageSend(m.ChannelID, "メモリタイプ:4gb")
					break
				} else {
					s.ChannelMessageSend(m.ChannelID, "プラン変更待機中...")
					time.Sleep(time.Minute)
				}
			}
			s.ChannelMessageSend(m.ChannelID, "プラン変更完了！これからサーバーを起動するよ！")
			time.Sleep(30 * time.Second)
			err := api.ServerCommand(token, "start")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			s.ChannelMessageSend(m.ChannelID, "サーバーを起動しました！")
		} else {
			s.ChannelMessageSend(m.ChannelID, "サーバーは既に起動しています")
		}
	} else if command == "!startsolo" {
		err := api.ServerCommand(token, "start")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}
		s.ChannelMessageSend(m.ChannelID, "サーバーを起動しました")

	} else if command == "!stop" {
		if serverStatus == "ACTIVE" {
			err := api.ServerCommand(token, "stop")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			s.ChannelMessageSend(m.ChannelID, "サーバーを停止しました")
		} else {
			s.ChannelMessageSend(m.ChannelID, "サーバーは既に停止しています")
			fmt.Println(flavor)
		}

		for {
			err := api.ChangeServerFlavor(token, flavor, "1gb")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			statusCheck, _ := api.GetServerStatus(token)
			if statusCheck == "VERIFY_RESIZE" {
				api.ConfirmResize(token)
				break
			} else {
				s.ChannelMessageSend(m.ChannelID, "プラン変更待機中...")
				time.Sleep(time.Minute)
			}
		}
		s.ChannelMessageSend(m.ChannelID, "プラン変更完了！")

	} else if command == "!reboot" {
		//再起動
		if serverStatus == "ACTIVE" {
			err := api.ServerCommand(token, "reboot")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			s.ChannelMessageSend(m.ChannelID, "サーバーを再起動しました")
		} else {
			s.ChannelMessageSend(m.ChannelID, "サーバーを起動して下さい")
		}

	} else if command == "!deposit" {
		deposit, err := api.GetPayment(token)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}
		s.ChannelMessageSend(m.ChannelID, strconv.Itoa(deposit)+"円の残高です")
	}
}

// func TimeSignal() (string, int) {
// 	now := time.Now()
// 	hour, min, _ := now.Clock()
// 	var message string
// 	if hour == 7 && min == 0 || hour == 19 && min == 0 {
// 		token := api.GetToken()
// 		serverStatus := api.GetServerStatus(token)
// 		if serverStatus == "SHUTOFF" {
// 			message = "サーバーは停止中です"
// 		} else if serverStatus == "ACTIVE" {
// 			message = "サーバーは起動中です"
// 		} else {
// 			message = "Unknown State"
// 		}
// 	}
// 	return message, hour
// }

// func Timer(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	for range time.Tick(1 * time.Minute) {
// 		// fmt.Println("受信")
// 		mes, hour := TimeSignal()
// 		if mes != "" && hour == 7 {
// 			s.ChannelMessageSend(m.ChannelID, "7時です！今日も一日頑張りましょう！")
// 			s.ChannelMessageSend(m.ChannelID, mes)
// 		} else if mes != "" && hour == 19 {
// 			s.ChannelMessageSend(m.ChannelID, "19時です！今日も一日お疲れさまでした！")
// 			s.ChannelMessageSend(m.ChannelID, mes)
// 		}
// 	}
// }

package rps

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/robfig/cron"
)

var payChannels, clientOpTimeout, clientModifyChannels = NewSynMap(), NewSynMap(), NewSynMap()

type compare func(interface{}, interface{}) bool

// Bot strcture.
type Bot struct {
	token       string
	opts        *Options
	crn         *cron.Cron
	users       *Users
	requests    *LDBMap
	stats       *LDBMap
	names       *LDBMap
	players     []int64
	leaderboard *[]*User
}

// New creates an object of Bot structure.
func New(
	token string,
	opts *Options,
	crn *cron.Cron,
	users *Users,
	requests *LDBMap,
	stats *LDBMap,
	names *LDBMap,
	players []int64,
	leaderboard *[]*User,
) Bot {
	b := Bot{token, opts, crn, users, requests, stats, names, players, leaderboard}
	return b
}

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("\U0001f39f BuyTicket"),
		tgbotapi.NewKeyboardButton("\U0001f4ec Subscribe"),
		tgbotapi.NewKeyboardButton("\U0001f3ad Change name"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("\U0001f5d1 Reset"),
		tgbotapi.NewKeyboardButton("\U0001f4ed Unsubscribe"),
		tgbotapi.NewKeyboardButton("\U0001f4b3 Change wallet address"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("\U00002753 Help"),
		tgbotapi.NewKeyboardButton("\U0001f50d Status"),
		tgbotapi.NewKeyboardButton("\U0001f3c6 Leaderboard"),
	),
)

var gameKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("\U000026f0 Rock"),
		tgbotapi.NewKeyboardButton("\U0001f4c4 Paper"),
		tgbotapi.NewKeyboardButton("\U00002702 Scissors"),
	),
)

func replyTo(
	chatID int64,
	reply string,
	botAPI *tgbotapi.BotAPI,
	markup interface{},
) {
	msg := tgbotapi.NewMessage(chatID, reply)
	//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	msg.ParseMode = "markdown"
	switch t := markup.(type) {
	default:
		Warning.Printf("Unexpected type %T", t)
	case tgbotapi.InlineKeyboardMarkup:
		msg.ReplyMarkup = markup.(tgbotapi.InlineKeyboardMarkup)
	case tgbotapi.ReplyKeyboardMarkup:
		msg.ReplyMarkup = markup.(tgbotapi.ReplyKeyboardMarkup)
	}
	if _, err := botAPI.Send(msg); err != nil {
		Error.Printf("Can't send reply to %d\n\t%s",
			chatID, err)
	}
}

func replyToMany(
	ids []int64,
	reply string,
	botAPI *tgbotapi.BotAPI,
	markup interface{},
) {
	for _, id := range ids {
		replyTo(id, reply, botAPI, markup)
	}
}

func clientOpTimeoutWatcher(chatID int64, opts *Options) {
	clientOpTimeout.Put(chatID, true)
	time.Sleep(time.Duration(opts.opTimeout) * time.Second)
	clientOpTimeout.Put(chatID, false)
}

func watchClientModify(chatID int64, channels *SynMap, opts *Options) string {
	var i uint
	ch := channels.Get(chatID).(chan string)
	defer close(ch)
	defer channels.Delete(chatID)

	Verbose.Printf("Watching for request to modify:\n\tChatID: %d",
		chatID)

	for {
		select {
		case val, _ := <-ch:
			if val == "" {
				Verbose.Printf("Watching for modify request has been reset:\n\tChatID: %d",
					chatID)
			} else {
				Verbose.Printf("Stopped to watch, data has been modified:\n\tChatID: %d",
					chatID)
			}
			return val
		default:
			if opts.modifyTime <= 5*i {
				Verbose.Printf("Stopped to watch for request to modify:\n\tChatID: %d",
					chatID)
				return ""
			}

			time.Sleep(time.Duration(5) * time.Second)
			i++
		}
	}
}

////////////****************************************************////////////
////////////***************** Bot methods start ****************////////////
////////////****************************************************////////////

// Welcome to everyone.
func (b *Bot) Welcome(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID

	reply = fmt.Sprintf(
		"*Hello and welcome to Rock-Paper-Scissors Online!*\n\n" +

			"Here you can play Rock-Paper-Scissors with others " +
			"as well as win some crypto currency. Rules are simple: " +
			"rock beat scissors, scissors beat paper and paper beat rock. " +
			"In case of two players choose the same item winner will be picked by random. " +
			"If you didn't make a move it will be made automatically by random pick of " +
			"rock, paper or scissors. Also there is special prize distribution: " +
			"the game is lost for you *only* if you're lost in the very first round " +
			"any other outcome is at least non-loss. For example, if you lose on the " +
			"second round you get your money back, if you lose on the " +
			"third round you get x2 of ticket price, if you lose on the " +
			"fourth round you get x3 of ticket price and so on. The final winner " +
			"get a special prize - all the non raffled money.\n\n" +

			"*Don't forget* to set up your wallet address otherwise your money " +
			"remain in the bank until somebody else win it!\n\n" +

			"*Commands you can use:*\n\n" +

			"/buyticket - buy a ticket\n" +
			"/reset - discard a payment request\n" +
			"/subscribe - subscribe onto the bot notifications\n" +
			"/unsubscribe - unsubscribe from the bot notifications\n" +
			"/status - current status of the game e.g. schedule, ticket price, etc.\n" +
			"/help - this message\n" +
			"/rock - make a move with rock\n" +
			"/paper - make a move with paper\n" +
			"/scissors - make a move with scissors\n" +
			"/leaderboard - show the leaderboard\n\n" +

			"*This bot doesn't take any of your money so the entire bank " +
			"pays out to players except Bitcoin Cash fees.*",
	)

	if b.opts.donationAddress != "" {
		reply += fmt.Sprintf("\n\nTo support this bot you can donate some coins to *%s* \U0000263a",
			b.opts.donationAddress)
	}

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// BuyTicket handles ticket purchase.
func (b *Bot) BuyTicket(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID
	replyError := func() {
		reply = "Something went wrong while processing request, please try again later."
		replyTo(chatID, reply, botAPI, mainKeyboard)
	}

	if !b.users.Exist(chatID) || b.users.Exist(chatID) && !b.users.Get(chatID).GetSubscribed() {
		needToBeSubscribed(chatID, botAPI)
		return
	}

	if b.requests.Exist(strconv.FormatInt(chatID, 10)) {
		reply = "You are in process of ticket purchase already."
		replyTo(chatID, reply, botAPI, mainKeyboard)
		return
	}

	if b.users.Exist(chatID) && b.users.Get(chatID).GetHasTicket() {
		reply = "You already have one!"
		replyTo(chatID, reply, botAPI, mainKeyboard)
		return
	}

	address, url, err := CreateRequest(b.opts.ticketPrice, b.opts.cashboxWalletPath, b.opts.testnet)
	if err != nil {
		Error.Printf("Can't create a new request:\n\t%s", err)
		replyError()
		return
	}
	if err := registerRequest(chatID, address, b.requests, &payChannels); err != nil {
		Error.Printf("Can't create a new request:\n\t%s", err)
		replyError()
		return
	}

	reply = fmt.Sprintf("*%s*", url)
	replyTo(chatID, reply, botAPI, mainKeyboard)
	reply = fmt.Sprintf("Okay, now you've got *%d minutes* to pay *%f BCH* to the address above.\n\n"+
		"If you wish to discard this request just type /reset or click to *Reset* button. "+
		"It's *NOT* recommended to reset paid transaction.",
		b.opts.payTime, b.opts.ticketPrice)
	replyTo(chatID, reply, botAPI, mainKeyboard)

	go b.processBuyTicket(chatID, &payChannels, botAPI)
}

// Reset resets ticket purchase and any active modifying actions.
func (b *Bot) Reset(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID

	if !b.users.Exist(chatID) || b.users.Exist(chatID) && !b.users.Get(chatID).GetSubscribed() {
		needToBeSubscribed(chatID, botAPI)
		return
	}

	replyTo(chatID, "Reseting in progress, please wait up to 15 seconds.", botAPI, mainKeyboard)

	if clientModifyChannels.Exist(chatID) {
		ch := clientModifyChannels.Get(chatID).(chan string)
		ch <- ""
	}

	if b.requests.Exist(strconv.FormatInt(chatID, 10)) {
		requestID := b.requests.Get(strconv.FormatInt(chatID, 10))
		if err := unregisterRequest(chatID, b.requests, &payChannels); err != nil {
			Warning.Printf("Can't unregister request:\n\tChatID: %d\n\t%s", chatID, err)
			reply = "Something went wrong, please try again."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		if err := RemoveRequest(requestID, b.opts.cashboxWalletPath, b.opts.testnet); err != nil {
			Error.Printf("Can't remove request:\n\tRequestID: %s\n\t%s", requestID, err)
		}
		Verbose.Printf("Reset successfully:\n\tChatID: %d\n\tRequestID: %s",
			chatID, requestID)
		reply = "Your payment request has been reset successfully."
	} else {
		Verbose.Printf("No transactions to reset for:\n\tChatID: %d", chatID)
		reply = "You have no transactions to reset."
	}

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// Subscribe enables notifications for user.
// Without subscription almost any action isn't available
func (b *Bot) Subscribe(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := "You're now subscribed!"
	chatID := update.Message.Chat.ID

	if b.users.Exist(chatID) && b.users.Get(chatID).GetSubscribed() {
		reply = "You're subscribed already."
	} else if b.users.Exist(chatID) && !b.users.Get(chatID).GetSubscribed() {
		user := b.users.Get(chatID)
		user.SetSubscribed(true)
		if err := b.users.Put(chatID, user); err != nil {
			Error.Printf("Can't subscribe:\n\tChatID: %d\n\t%s",
				chatID, err)
			reply = "Something went wrong, please try again."
		}
	} else {
		name := update.Message.Chat.UserName
		for name == "" || b.names.Exist(name) || !NameValidate(name) {
			name = randomdata.SillyName()
		}
		if err := b.names.Put(name, ""); err != nil {
			Error.Printf("Can't save generated name:\n\tChatID: %d\n\tName: %s",
				chatID, name)
			reply = "Something went wrong, please try again."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		user := NewUser(chatID, name, true, false, false, uint32(b.users.Len()+1))
		if err := b.users.Put(chatID, &user); err != nil {
			Error.Printf("Can't subscribe:\n\tChatID: %d\n\t%s",
				chatID, err)
			reply = "Something went wrong, please try again."
		}
	}

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// Unsubscribe disables notifications for user.
func (b *Bot) Unsubscribe(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	var markup tgbotapi.InlineKeyboardMarkup
	chatID := update.Message.Chat.ID

	if b.users.Exist(chatID) && b.users.Get(chatID).GetHasTicket() {
		reply := "You'll lose your ticket. Are you sure you want to unsubscribe?"
		yes := tgbotapi.NewInlineKeyboardButtonData("Yes", "yes")
		no := tgbotapi.NewInlineKeyboardButtonData("No", "no")
		markup.InlineKeyboard = append(markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			yes, no,
		))
		replyTo(chatID, reply, botAPI, markup)
	} else {
		b.YesUnsubscribe(update, botAPI)
	}
}

func needToBeSubscribed(chatID int64, botAPI *tgbotapi.BotAPI) {
	reply := "To perform this operation you need to subscribe."
	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// Status shows status message filled up with user's stats.
func (b *Bot) Status(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	var user *User
	reply := ""
	chatID := update.Message.Chat.ID

	if b.users.Exist(chatID) {
		user = b.users.Get(chatID)
	} else {
		needToBeSubscribed(chatID, botAPI)
		return
	}

	if !user.GetSubscribed() {
		needToBeSubscribed(chatID, botAPI)
		return
	}

	reply = fmt.Sprintf("_%s_ (_%d_)\n", user.GetName(), user.GetUserID())

	reply += fmt.Sprintf(
		"\n\U0001f48e Ticket price: *%0.3f BTH*\n\U0001f551 Next game launch: *%s*",
		b.opts.ticketPrice,
		b.crn.Entries()[0].Next.Format(time.RFC1123),
	)

	if user.GetHasTicket() {
		reply += "\n\U0001f3b2 You *have* a ticket"
	} else {
		reply += "\n\U0001f614 You *have no* ticket"
	}

	reply += fmt.Sprintf("\n\U0001f4b3 Your wallet address: *%s*", user.GetWalletAddress())

	reply += fmt.Sprintf("\n\U0001f4b0 Your total won amount: *%f*", user.GetTotalWonAmount())

	reply += fmt.Sprintf("\n\U0001f3c5 Your position in the leaderboard is *%d* of *%d*",
		user.GetLeaderboardPosition(), b.users.Len())

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// Leaderboard shows leaderboard message.
func (b *Bot) Leaderboard(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID

	if b.users.Exist(chatID) {
		if len(*b.leaderboard) > 0 {
			for i, el := range (*b.leaderboard)[:Min(10, len(*b.leaderboard))] {
				reply += fmt.Sprintf("%d. %s\t\t%f\n", i+1, el.GetName(), el.GetTotalWonAmount())
			}
		} else {
			reply = "Leaderboard is empty yet."
		}
		replyTo(chatID, reply, botAPI, mainKeyboard)
	} else {
		needToBeSubscribed(chatID, botAPI)
	}
}

// ChangeName updates username of user.
func (b *Bot) ChangeName(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID
	replyError := func() {
		reply = "Something went wrong, please try again."
		replyTo(chatID, reply, botAPI, mainKeyboard)
	}

	if b.users.Exist(chatID) && b.users.Get(chatID).GetSubscribed() {
		if clientModifyChannels.Exist(chatID) {
			reply = "You're already in process of modifying your data. You can reset it by /reset."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		user := b.users.Get(chatID)
		reply = "Enter new username please."
		replyTo(chatID, reply, botAPI, mainKeyboard)

		ch := make(chan string)
		clientModifyChannels.Put(chatID, ch)

		name := watchClientModify(chatID, &clientModifyChannels, b.opts)
		if name == "" {
			reply = "Request is expired or reset."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		if NameValidate(name) && !b.names.Exist(name) {
			oldName := user.GetName()
			user.SetName(name)
			if err := b.names.Delete(oldName); err != nil {
				Error.Printf("Can't delete old name:\n\tChatID: %d\n\tName: %s",
					chatID, name)
				replyError()
				return
			}
			if err := b.users.Put(chatID, user); err != nil {
				Error.Printf("Can't save user with new name:\n\tChatID: %d\n\tName: %s",
					chatID, name)
				replyError()
				return
			}
			if err := b.names.Put(name, ""); err != nil {
				Error.Printf("Can't save new name:\n\tChatID: %d\n\tName: %s",
					chatID, name)
				replyError()
				return
			}
			reply = "Username set successfully!"
		} else {
			reply = "This username isn't valid or occupied by someone else, try to change something."
		}
		replyTo(chatID, reply, botAPI, mainKeyboard)
	} else {
		needToBeSubscribed(chatID, botAPI)
	}
}

// ChangeWalletAddress updates wallet address of user.
func (b *Bot) ChangeWalletAddress(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply := ""
	chatID := update.Message.Chat.ID

	if b.users.Exist(chatID) && b.users.Get(chatID).GetSubscribed() {
		if clientModifyChannels.Exist(chatID) {
			reply = "You're already in process of modifying your data. You can reset it by /reset."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		user := b.users.Get(chatID)
		reply = "Enter new wallet address please."
		replyTo(chatID, reply, botAPI, mainKeyboard)

		ch := make(chan string)
		clientModifyChannels.Put(chatID, ch)

		wallet := watchClientModify(chatID, &clientModifyChannels, b.opts)
		if wallet == "" {
			reply = "Request is expired or reset."
			replyTo(chatID, reply, botAPI, mainKeyboard)
			return
		}
		if WalletValidate(wallet) {
			user.SetWalletAddress(wallet)
			b.users.Put(chatID, user)
			reply = "Wallet set successfully!"
		} else {
			reply = "This wallet isn't valid, try to change something. " +
				"Note that you need to input in a cash address format, not a legacy one."
		}
		replyTo(chatID, reply, botAPI, mainKeyboard)
	} else {
		needToBeSubscribed(chatID, botAPI)
	}
}

////////////****************************************************////////////
////////////***************** Bot methods end ******************////////////
////////////****************************************************////////////

////////////****************************************************////////////
////////////************ Inline callbacks start ****************////////////
////////////****************************************************////////////

// YesUnsubscribe confirms unsubscribe action.
// Appear only when user has ticket.
func (b *Bot) YesUnsubscribe(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	var chatID int64
	reply := "You're now unsubscribed."

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}

	if !b.users.Exist(chatID) || b.users.Exist(chatID) && !b.users.Get(chatID).GetSubscribed() {
		reply = "You're not subscribed."
		replyTo(chatID, reply, botAPI, mainKeyboard)
		return
	}
	user := b.users.Get(chatID)
	user.SetSubscribed(false)
	user.SetHasTicket(false)
	if err := b.users.Put(chatID, user); err != nil {
		Error.Printf("Can't unsubscribe:\n\tChatID: %d\n\t%s",
			chatID, err)
		reply = "Something went wrong, pleaase try again."
	}

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

// NoUnsubscribe dismiss unsubscribe action.
func (b *Bot) NoUnsubscribe(update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	chatID := update.CallbackQuery.Message.Chat.ID
	reply := "You won't be unsubscribed."

	if !b.users.Exist(chatID) || b.users.Exist(chatID) && !b.users.Get(chatID).GetSubscribed() {
		reply = "You're not subscribed."
	}

	replyTo(chatID, reply, botAPI, mainKeyboard)
}

////////////****************************************************////////////
////////////************ Inline callbacks end ******************////////////
////////////****************************************************////////////

////////////****************************************************////////////
////////////************ Wallet operations start ***************////////////
////////////****************************************************////////////

func (b *Bot) cleanupProcessBuyTicket(
	chatID int64,
	requestID string,
	channels *SynMap,
	botAPI *tgbotapi.BotAPI,
) {
	Verbose.Printf("Cleaning up request for:\n\tChatID: %d\n\tRequestID: %s",
		chatID, requestID)

	if err := unregisterRequest(chatID, b.requests, channels); err != nil {
		Warning.Printf("Can't unregister request:\n\tChatID: %d\n\t%s", chatID, err)
	}

	if err := RemoveRequest(requestID, b.opts.cashboxWalletPath, b.opts.testnet); err != nil {
		Error.Printf("Can't remove request:\n\tRequestID: %s\n\t%s", requestID, err)
	}
}

func (b *Bot) processBuyTicket(
	chatID int64,
	channels *SynMap,
	botAPI *tgbotapi.BotAPI,
) {
	var paymentStatus uint8
	reply := ""
	requestID := b.requests.Get(strconv.FormatInt(chatID, 10))

	paymentStatus, err := b.processRequest(chatID, requestID, channels, botAPI)
	if err != nil {
		Error.Printf("Request can't be processed:\n\tChatID: %d\n\tRequestID: %s\n\t%s",
			chatID, requestID, err)
		reply = "Can't process your request, please try again."
		replyTo(chatID, reply, botAPI, mainKeyboard)
		b.cleanupProcessBuyTicket(chatID, requestID, channels, botAPI)
		return
	}
	if paymentStatus == 0 {
		Verbose.Printf("Successfully got \"Paid\" status for:\n\tChatID: %d\n\tRequestID: %s",
			chatID, requestID)

		user := b.users.Get(chatID)
		user.SetHasTicket(true)
		user.SetLastTicketDate(time.Now())
		if err := b.users.Put(chatID, user); err != nil {
			Error.Printf("Can't give a ticket to the player:\n\tChatID: %d\n\tRequestID: %s",
				chatID, requestID)
		}

		reply = "You've got a ticket \U0001f39f To check current game schedule type /status."
		replyTo(chatID, reply, botAPI, mainKeyboard)
		b.cleanupProcessBuyTicket(chatID, requestID, channels, botAPI)
	} else if paymentStatus == 1 {
		Verbose.Printf("Time is up for:\n\tChatID: %d\n\tRequestID: %s",
			chatID, requestID)
		reply = "Time is up, would you like to try to /buyticket again?"
		replyTo(chatID, reply, botAPI, mainKeyboard)
		b.cleanupProcessBuyTicket(chatID, requestID, channels, botAPI)
	} else {
		Verbose.Printf("Transaction has been reset:\n\tChatID: %d\n\tequestID:%s",
			chatID, requestID)
	}
}

func registerRequest(
	chatID int64,
	address string,
	requests *LDBMap,
	channels *SynMap,
) error {
	err := requests.Put(strconv.FormatInt(chatID, 10), address)
	if err != nil {
		return err
	}
	ch := make(chan bool)
	channels.Put(chatID, ch)

	return nil
}

func unregisterRequest(
	chatID int64,
	requests *LDBMap,
	channels *SynMap,
) error {
	if err := requests.Delete(strconv.FormatInt(chatID, 10)); err != nil {
		return err
	}

	if channels.Exist(chatID) {
		ch := channels.Get(chatID).(chan bool)
		close(ch)
		channels.Delete(chatID)
	}

	return nil
}

func (b *Bot) processRequest(
	chatID int64,
	requestID string,
	channels *SynMap,
	botAPI *tgbotapi.BotAPI,
) (uint8, error) {
	paymentStatus, err := b.watchTransaction(chatID,
		"status", "Paid", channels,
		func(a interface{}, b interface{}) bool {
			if a == b {
				return true
			}
			return false
		},
	)

	if err != nil {
		return paymentStatus, err
	}

	return paymentStatus, nil
}

func (b *Bot) watchTransaction(
	chatID int64,
	key string,
	value interface{},
	channels *SynMap,
	cmp compare,
) (uint8, error) {
	var requestField interface{}
	var i uint
	requestID := b.requests.Get(strconv.FormatInt(chatID, 10))
	ch := channels.Get(chatID).(chan bool)

	Verbose.Printf("Watching for request:\n\tChatID: %d\n\tRequestID: %s",
		chatID, requestID)

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				Verbose.Printf("Watching has been reset:\n\tChatID: %d\n\tRequestID: %s",
					chatID, requestID)
				return 2, nil
			}
		default:
			request, err := GetRequest(requestID, b.opts.cashboxWalletPath, b.opts.testnet)
			if err != nil {
				Error.Printf("Can't get request:\n\tRequestID: %s\n\t%s", requestID, err)
				return 1, err
			}
			json.Unmarshal(request[key], &requestField)

			if cmp(requestField, value) {
				Verbose.Printf("Stopped to watch for request:\n\tChatID: %d\n\tRequestID: %s",
					chatID, requestID)
				return 0, nil
			}

			if b.opts.payTime*60 <= i*5 {
				Verbose.Printf("Stopped to watch for request:\n\tChatID: %d\n\tRequestID: %s",
					chatID, requestID)
				return 1, nil
			}

			time.Sleep(time.Duration(5) * time.Second)
			i++
		}
	}
}

////////////****************************************************////////////
////////////************ Wallet operations end *****************////////////
////////////****************************************************////////////

////////////****************************************************////////////
////////////*************** Game methods start *****************////////////
////////////****************************************************////////////

func formLeaderboard(users *Users) []*User {
	totalWonAmountLST := users.FormTotalWonAmountList()
	leaderboard := make([]*User, len(totalWonAmountLST))
	for i := len(totalWonAmountLST) - 1; i >= 0; i-- {
		leaderboard[len(totalWonAmountLST)-i-1] = totalWonAmountLST[i]
		totalWonAmountLST[i].SetLeaderboardPosition(uint32(len(totalWonAmountLST) - i))
		users.BatchPut(totalWonAmountLST[i].GetUserID(), totalWonAmountLST[i])
	}
	if err := users.BatchWrite(); err != nil {
		Error.Printf("Can't update leaderboard positions.")
	}

	return leaderboard
}

func alignPlayers(players []int64, opts *Options) ([]int64, []int64) {
	var c uint = 2
	for c < opts.capacity {
		if c > uint(len(players)) {
			return players[:c>>1], players[c>>1:]
		}
		c = c << 1
	}

	return players[:opts.capacity], players[opts.capacity:]
}

func transitionToReady(stats *LDBMap) error {
	if err := stats.Put("ready", "true"); err != nil {
		return err
	}

	return nil
}

func transitionToGame(stats *LDBMap) error {
	if err := stats.Put("ready", "false"); err != nil {
		return err
	}
	if err := stats.Put("game", "true"); err != nil {
		return err
	}

	return nil
}

// GamePrepare takes all necessary actions to prepare the game.
// Sort users by last ticket purchase date and align to ^2 number.
func (b *Bot) GamePrepare(botAPI *tgbotapi.BotAPI) {
	reply := ""
	lastTicketDateSorted := b.users.FormLastTicketDateList()
	replyCritical := func() {
		reply = "Due the critical error game couldn't start this time, please " +
			"accept our apologies and wait for the next round." +
			"Your funds are probably safe and sound \U0001f642"
		replyToMany(b.players, reply, botAPI, mainKeyboard)
	}

	if err := b.stats.Put("prepare", "true"); err != nil {
		Error.Printf("Can't start preparation stage. Can't set prepare to true\n\t%s", err)
		replyCritical()
		return
	}

	if b.stats.Get("ready") == "false" {
		for _, user := range lastTicketDateSorted {
			if user.GetHasTicket() == true {
				userID := user.GetUserID()
				b.players = append(b.players, userID)
			}
		}

		if len(b.players) < 2 {
			Info.Printf("Not enough players, game won't start.")
			reply = "There is not enough players, can't start the game for now."
			replyToMany(b.players, reply, botAPI, mainKeyboard)
			b.players = []int64{}
			return
		}

		tail := []int64{}
		b.players, tail = alignPlayers(b.players, b.opts)
		for _, chatID := range b.players {
			user := b.users.Get(chatID)
			user.SetIsPlayer(true)
			user.SetPlaySequence("")
			user.SetLastWonAmount(0)
			b.users.BatchPut(chatID, user)
			reply = fmt.Sprintf("Get ready, game is starting! This time %d players are taking a part.",
				len(b.players))
			replyTo(chatID, reply, botAPI, mainKeyboard)
		}

		reply = "Game is crowded for now, your ticket will play next round."
		replyToMany(tail, reply, botAPI, mainKeyboard)

		if err := b.users.BatchWrite(); err != nil {
			Error.Printf("Can't prepare users to the game.")
			replyToMany(b.players, reply, botAPI, mainKeyboard)
			return
		}

		address, _, err := CreateRequest(1, b.opts.bankWalletPath, b.opts.testnet)
		if err != nil {
			Error.Printf("Can't create request to move money to the bank. CRITICAL.\n\t%s", err)
			replyCritical()
			return
		}
		Info.Printf("Request to move funds to the bank created successfully.")
		move := -1.0
		if len(tail) != 0 {
			move = float64(len(b.players)) * b.opts.ticketPrice
		}
		if err := PayTo(address, move, b.opts.cashboxWalletPath, b.opts.testnet); err != nil {
			Error.Printf("Can't move money to the bank. CRITICAL.\n\t%s", err)
			replyCritical()
			return
		}
		Info.Printf("Funds have been moved to the bank successfully.")
		if err := ClearRequests(b.opts.bankWalletPath, b.opts.testnet); err != nil {
			Warning.Printf("Can't clear requests of the bank wallet:\n\t%s", err)
		} else {
			Verbose.Printf("Requests of the bank wallet cleared successfully.")
		}

		// Set ready status to true in case of server shutdown before the game start
		if err := transitionToReady(b.stats); err != nil {
			Error.Printf("Can't make a transition to the prepare stage\n\t%s", err)
			replyCritical()
			return
		}
	}

	// Wait 10 seconds before start the actual game
	time.Sleep(10 * time.Second)

	// Set game status to true in case of server shutdown before the game end
	if err := transitionToGame(b.stats); err != nil {
		Error.Printf("Can't make a transition to the game stage\n\t%s", err)
		replyCritical()
		return
	}

	go b.Play(botAPI)
}

// GameRestore resurects game if bot crashed.
func (b *Bot) GameRestore(botAPI *tgbotapi.BotAPI) {
	for uid, user := range b.users.Iterate() {
		if user.GetIsPlayer() == true {
			b.players = append(b.players, uid)
		}
	}
	tail := []int64{}
	b.players, tail = alignPlayers(b.players, b.opts)
	reply := "Something wrong has happened, sorry for inconvenience. The game continues!"
	replyToMany(b.players, reply, botAPI, gameKeyboard)
	for _, id := range tail {
		userReset(id, b.users)
		user := b.users.Get(id)
		if err := PayToUser(user, user.GetLastWonAmount(), b.opts.bankWalletPath, b.opts.testnet); err != nil {
			Error.Printf("Couldn't pay to user:\n\tUserID: %d\n\tUsername: %s\n\t%s",
				user.GetUserID(), user.GetName(), err)
		}
		reply = fmt.Sprintf("Something wrong has happened, very sorry for inconvenience, "+
			"but this game is ended for you \U0001f614 Won amount: *%f BCH* \U0001f4b6",
			user.GetLastWonAmount())
		replyTo(id, reply, botAPI, mainKeyboard)
	}

	reply = "Due the critical error game couldn't start this time, please " +
		"accept our apologies and wait for the next round. Your funds are probably safe and sound \U0001f642"
	if err := transitionToGame(b.stats); err != nil {
		Error.Printf("Can't make a transition to the game\n\t%s", err)
		replyToMany(b.players, reply, botAPI, mainKeyboard)
		return
	}

	go b.Play(botAPI)
}

// GameReset resets all users to pregame state and cleans players list.
func (b *Bot) GameReset(botAPI *tgbotapi.BotAPI) {
	for _, chatID := range b.players {
		userReset(chatID, b.users)
		replyTo(chatID, "This round is over, thank you for the game!", botAPI, mainKeyboard)
	}

	b.players = make([]int64, 0)
	if err := b.stats.Put("game", "false"); err != nil {
		Error.Printf("Can't set game status to false\n\t%s", err)
	} else {
		Info.Printf("Game reset successfully")
	}
}

func userReset(id int64, users *Users) {
	user := users.Get(id)
	user.SetIsPlayer(false)
	user.SetHasTicket(false)
	user.SetPlaySequence("")
	if err := users.Put(id, user); err != nil {
		Error.Printf("Can't reset the user\n\tUserID: %d\n\tUsername: %s\n\t%s",
			id, user.GetName(), err)
	} else {
		Verbose.Printf("User has been reset\n\tUserID: %d\n\tUsername: %s",
			id, user.GetName())
	}
}

// MakeAMove implements user's move.
func (b *Bot) MakeAMove(move byte, update tgbotapi.Update, botAPI *tgbotapi.BotAPI) {
	reply, moves := "", ""
	chatID := update.Message.Chat.ID

	if b.stats.Get("game") == "false" {
		reply = fmt.Sprintf("There is no game in process. To see the schedule " +
			"plase use /status command or just tap to the Status button.")
		replyTo(chatID, reply, botAPI, mainKeyboard)
		return
	}

	user := b.users.Get(chatID)
	moves = user.GetPlaySequence()
	if len(moves) > 0 && moves[len(moves)-1] == '#' {
		moves = moves[:len(moves)-2]
	}
	moves += string(move) + "#"
	user.SetPlaySequence(moves)
	if err := b.users.Put(chatID, user); err != nil {
		Error.Printf("Can't put an element to the player's moves bucket\n\t%s", err)
	}

	reply = fmt.Sprintf("Your moves for now: %s*%s*",
		moves[:len(moves)-2], string(moves[len(moves)-2]))
	replyTo(chatID, reply, botAPI, gameKeyboard)
}

func round(
	playerA, playerB int64,
	users *Users,
	ch chan int64,
	wg *sync.WaitGroup,
	opts *Options,
	botAPI *tgbotapi.BotAPI,
) {
	defer wg.Done()
	var winner, loser int64
	reply := ""

	if playerA == -1 || playerB == -1 {
		if playerA == -1 {
			ch <- playerB
			ch <- playerA
		} else {
			ch <- playerA
			ch <- playerB
		}
		return
	}

	userA, userB := users.Get(playerA), users.Get(playerB)
	playerASequence := userA.GetPlaySequence()
	playerBSequence := userB.GetPlaySequence()
	draw := rand.Intn(2)

	reply = fmt.Sprintf("You have %d second to make a move.", opts.roundTime)
	if playerBSequence != "" {
		reply = fmt.Sprintf("Opponent's sequence: %s*%s*\nYou have %d second to make a move.",
			playerBSequence[:len(playerBSequence)-1],
			string(playerBSequence[len(playerBSequence)-1]),
			opts.roundTime)
	}
	replyTo(playerA, reply, botAPI, gameKeyboard)
	reply = fmt.Sprintf("You have %d second to make a move.", opts.roundTime)
	if playerASequence != "" {
		reply = fmt.Sprintf("Opponent's sequence: %s*%s*\nYou have %d second to make a move.",
			playerASequence[:len(playerASequence)-1],
			string(playerASequence[len(playerASequence)-1]),
			opts.roundTime)
	}
	replyTo(playerB, reply, botAPI, gameKeyboard)

	// Timeout to let players make a move
	time.Sleep(time.Duration(opts.roundTime) * time.Second)

	if draw == 0 {
		winner = playerA
		loser = playerB
	} else {
		winner = playerB
		loser = playerA
	}

	// Refresh player's sequences after the turn
	playerASequence = userA.GetPlaySequence()
	playerBSequence = userB.GetPlaySequence()

	rps := []string{"R", "P", "S"}
	if len(playerASequence) == 0 || playerASequence[len(playerASequence)-1] != '#' {
		r1 := rand.Intn(len(rps))
		playerASequence += rps[r1] + "#"
		userA.SetPlaySequence(playerASequence)
		if err := users.Put(playerA, userA); err != nil {
			Error.Printf("Can't put updated user A play sequence\n\t%s", err)
		}
		reply = fmt.Sprintf("Your moves for now: %s*%s*",
			playerASequence[:len(playerASequence)-2],
			string(playerASequence[len(playerASequence)-2]))
		replyTo(playerA, reply, botAPI, gameKeyboard)
	}
	if len(playerBSequence) == 0 || playerBSequence[len(playerBSequence)-1] != '#' {
		r2 := rand.Intn(len(rps))
		playerBSequence += rps[r2] + "#"
		userB.SetPlaySequence(playerBSequence)
		if err := users.Put(playerB, userB); err != nil {
			Error.Printf("Can't put updated user B play sequence\n\t%s", err)
		}
		reply = fmt.Sprintf("Your moves for now: %s*%s*",
			playerBSequence[:len(playerBSequence)-2],
			string(playerBSequence[len(playerBSequence)-2]))
		replyTo(playerB, reply, botAPI, gameKeyboard)
	}

	reply = fmt.Sprintf("Opponent's move: *%s*", string(playerBSequence[len(playerBSequence)-2]))
	replyTo(playerA, reply, botAPI, gameKeyboard)
	reply = fmt.Sprintf("Opponent's move: *%s*", string(playerASequence[len(playerASequence)-2]))
	replyTo(playerB, reply, botAPI, gameKeyboard)

	if playerASequence[len(playerASequence)-2] == 'R' {
		switch playerBSequence[len(playerASequence)-2] {
		case 'P':
			winner = playerB
			loser = playerA
		case 'S':
			winner = playerA
			loser = playerB
		}
	} else if playerASequence[len(playerASequence)-2] == 'P' {
		switch playerBSequence[len(playerASequence)-2] {
		case 'R':
			winner = playerA
			loser = playerB
		case 'S':
			winner = playerB
			loser = playerA
		}
	} else if playerASequence[len(playerASequence)-2] == 'S' {
		switch playerBSequence[len(playerASequence)-2] {
		case 'R':
			winner = playerB
			loser = playerA
		case 'P':
			winner = playerA
			loser = playerB
		}
	}

	ch <- winner
	ch <- loser
}

// Play starts the game.
func (b *Bot) Play(botAPI *tgbotapi.BotAPI) {
	Info.Printf("Game of %d players is starting", len(b.players))
	reply := ""

	for len(b.players) > 1 {
		gameChannels := NewSynMap()

		// Pause in-between rounds
		time.Sleep(time.Duration(b.opts.timeout) * time.Second)

		rand.Shuffle(len(b.players),
			func(i, j int) { b.players[i], b.players[j] = b.players[j], b.players[i] })

		var wg sync.WaitGroup
		wg.Add(len(b.players) / 2)

		for i := 0; i < len(b.players); i += 2 {
			ch := make(chan int64, 2)
			gameChannels.Put(i, ch)
			go round(b.players[i], b.players[i+1], b.users, ch, &wg, b.opts, botAPI)
		}
		wg.Wait()
		for _, ch := range gameChannels.Iterate() {
			winner, loser := <-ch.(chan int64), <-ch.(chan int64)
			idx := ContainsInt64(loser, b.players)
			if idx != -1 {
				b.players = append(b.players[:idx], b.players[idx+1:]...)
			}

			userLoser := b.users.Get(loser)
			userReset(loser, b.users)
			userWinner := b.users.Get(winner)

			userWinner.SetLastWonAmount(userWinner.GetLastWonAmount() + b.opts.ticketPrice)
			if len(b.players) == 1 {
				reply = fmt.Sprintf("You won the final prize \U0001f389 "+
					"Won amount: *%f BCH* plus extra coins \U0001f381", userWinner.GetLastWonAmount())
				if b.opts.donationAddress != "" {
					reply += fmt.Sprintf(" You can support this bot by donating to *%s* "+
						"Thank you and have a nice day \U0001f60a", b.opts.donationAddress)
				}
				replyTo(winner, reply, botAPI, gameKeyboard)
				if err := PayToUser(userWinner, -1, b.opts.bankWalletPath, b.opts.testnet); err != nil {
					Error.Printf("Couldn't pay to user:\n\tUserID: %d\n\tUsername: %s\n\t%s",
						userWinner.GetUserID(), userWinner.GetName(), err)
				}
				Info.Printf("Final winner:\n\tUserID: %d\n\tUsername: %s\n\tAmount: %f",
					userWinner.GetUserID(), userWinner.GetName(), userWinner.GetLastWonAmount())
			} else {
				reply = fmt.Sprintf("You win! Won amount: *%f BCH* \U0001f4b6",
					userWinner.GetLastWonAmount())
				replyTo(winner, reply, botAPI, gameKeyboard)
				Info.Printf("Winner:\n\tUserID: %d\n\tUsername: %s\n\tAmount: %f",
					userWinner.GetUserID(), userWinner.GetName(), userWinner.GetLastWonAmount())
			}

			reply = fmt.Sprintf("You lose! Won amount: *%f BCH* \U0001f4b6",
				userLoser.GetLastWonAmount())
			if userLoser.GetLastWonAmount() > 0.0 {
				if err := PayToUser(userLoser, userLoser.GetLastWonAmount(),
					b.opts.bankWalletPath, b.opts.testnet); err != nil {
					Error.Printf("Couldn't pay to user:\n\tUserID: %d\n\tUsername: %s\n\t%s",
						userLoser.GetUserID(), userLoser.GetName(), err)
				}
				if userLoser.GetLastWonAmount() > b.opts.ticketPrice*3 &&
					b.opts.donationAddress != "" {
					reply += fmt.Sprintf(" \n\nYou can support this bot by donating to *%s* "+
						"Thank you and have a nice day \U0001f60a", b.opts.donationAddress)
				}
			}
			replyTo(loser, reply, botAPI, mainKeyboard)
			Info.Printf("Loser:\n\tUserID: %d\n\tUsername: %s\n\tAmount: %f",
				userLoser.GetUserID(), userLoser.GetName(), userLoser.GetLastWonAmount())

			userWinner.SetTotalWonAmount(userWinner.GetTotalWonAmount() +
				userWinner.GetLastWonAmount())
			if err := b.users.Put(winner, userWinner); err != nil {
				Error.Printf("Can't update winner after the round\n\t%s", err)
			}
		}

		for _, id := range b.players {
			user := b.users.Get(id)
			moves := user.GetPlaySequence()
			if len(moves) > 0 && moves[len(moves)-1] == '#' {
				moves = moves[:len(moves)-1]
				user.SetPlaySequence(moves)
				b.users.BatchPut(id, user)
			}
		}
		if err := b.users.BatchWrite(); err != nil {
			Error.Printf("Can't remove terminal symbol from play sequences.")
		}
	}

	b.GameReset(botAPI)
	*b.leaderboard = []*User{}
	for _, u := range formLeaderboard(b.users) {
		*b.leaderboard = append(*b.leaderboard, u)
	}
}

////////////****************************************************////////////
////////////*************** Game methods end *******************////////////
////////////****************************************************////////////

// Start starts the bot.
func (b *Bot) Start() {
	botAPI, err := tgbotapi.NewBotAPI(b.token)
	if err != nil {
		Error.Println("Can't authenticate with given token")
		panic(err)
	}

	Info.Printf("Authorized on account %s", botAPI.Self.UserName)

	rand.Seed(time.Now().Unix())

	Verbose.Printf("Restoring interrupted requests...")
	*b.requests = NewLDBMap("requests", b.opts.dbPath)
	defer b.requests.Close()
	for k := range b.requests.Iterate() {
		chatID, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			Error.Printf("Can't parse user ID: %s", err)
		}
		ch := make(chan bool)
		payChannels.Put(chatID, ch)
		go b.processBuyTicket(chatID, &payChannels, botAPI)
		reply := "Something wrong has happened, sorry for inconvenience. " +
			"Service just restarted and you can continue with your payment process."
		replyTo(chatID, reply, botAPI, mainKeyboard)
	}
	Verbose.Printf("%d interrupted requests restored", b.requests.Len())

	Verbose.Printf("Loading users...")
	*b.users = NewUsers("users", b.opts.dbPath)
	Verbose.Printf("%d users loaded", b.users.Len())

	Verbose.Printf("Loading used names...")
	*b.names = NewLDBMap("names", b.opts.dbPath)
	Verbose.Printf("%d used names loaded", b.names.Len())

	*b.stats = NewLDBMap("stats", b.opts.dbPath)
	defer b.stats.Close()
	if b.stats.Get("game") == "true" || b.stats.Get("ready") == "true" {
		b.GameRestore(botAPI)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := botAPI.GetUpdatesChan(u)

	b.crn.AddFunc(b.opts.schedule, func() {
		if b.stats.Get("game") != "true" {
			b.GamePrepare(botAPI)
		}
	})

	// From leaderboard
	*b.leaderboard = []*User{}
	for _, u := range formLeaderboard(b.users) {
		*b.leaderboard = append(*b.leaderboard, u)
	}

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			if clientOpTimeout.Exist(chatID) && clientOpTimeout.Get(chatID).(bool) {
				reply := fmt.Sprintf(
					"Please wait a little before calling again :) Timeout is equal to %d seconds.",
					b.opts.opTimeout,
				)
				idx := ContainsInt64(chatID, b.players)
				if idx != -1 {
					replyTo(chatID, reply, botAPI, gameKeyboard)
				} else {
					replyTo(chatID, reply, botAPI, mainKeyboard)
				}
				continue
			}
			go clientOpTimeoutWatcher(chatID, b.opts)

			Info.Printf("[%d] %s", chatID, update.Message.Text)

			switch update.Message.Text {
			case "/start", "start", "Start":
				go b.Welcome(update, botAPI)
				go b.Subscribe(update, botAPI)
			case "/buyticket", "buyticket", "BuyTicket", "\U0001f39f BuyTicket":
				go b.BuyTicket(update, botAPI)
			case "/reset", "reset", "Reset", "\U0001f5d1 Reset":
				go b.Reset(update, botAPI)
			case "/subscribe", "subscribe", "Subscribe", "\U0001f4ec Subscribe":
				go b.Subscribe(update, botAPI)
			case "/unsubscribe", "unsubscribe", "Unsubscribe", "\U0001f4ed Unsubscribe":
				go b.Unsubscribe(update, botAPI)
			case "/status", "status", "Status", "\U0001f50d Status":
				go b.Status(update, botAPI)
			case "/changename", "change name", "Change name", "\U0001f3ad Change name":
				go b.ChangeName(update, botAPI)
			case "/changewalletaddress", "change wallet address",
				"Change wallet address", "\U0001f4b3 Change wallet address":
				go b.ChangeWalletAddress(update, botAPI)
			case "/leaderboard", "leaderboard", "Leaderboard", "\U0001f3c6 Leaderboard":
				go b.Leaderboard(update, botAPI)
			case "/rock", "rock", "Rock", "\U000026f0 Rock":
				go b.MakeAMove('R', update, botAPI)
			case "/paper", "paper", "Paper", "\U0001f4c4 Paper":
				go b.MakeAMove('P', update, botAPI)
			case "/scissors", "scissors", "Scissors", "\U00002702 Scissors":
				go b.MakeAMove('S', update, botAPI)
			case "/help", "help", "Help", "\U00002753 Help":
				go b.Welcome(update, botAPI)
			default:
				if clientModifyChannels.Exist(chatID) {
					ch := clientModifyChannels.Get(chatID).(chan string)
					ch <- update.Message.Text
				}
			}
		}

		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			if clientOpTimeout.Exist(chatID) && clientOpTimeout.Get(chatID).(bool) {
				reply := fmt.Sprintf(
					"Please wait a little before calling again :) Timeout is equal to %d seconds.",
					b.opts.opTimeout,
				)
				replyTo(chatID, reply, botAPI, mainKeyboard)
				continue
			}
			go clientOpTimeoutWatcher(chatID, b.opts)

			Info.Printf("[%d] %s", chatID, update.CallbackQuery.Data)

			switch update.CallbackQuery.Data {
			case "yes":
				go b.YesUnsubscribe(update, botAPI)
			case "no":
				go b.NoUnsubscribe(update, botAPI)
			}
		}
	}
}

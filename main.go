package main

import (
	"io/ioutil"
	"os"

	"github.com/robfig/cron"
	"github.com/rps-bot/rpsbot/rps"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	capacity = kingpin.Flag(
		"capacity",
		"Capacity of each game.",
	).Default("128").Short('c').Uint()
	timeout = kingpin.Flag(
		"timeout",
		"Timeout in-between of rounds (in seconds).",
	).Default("5").Short('t').Uint()
	opTimeout = kingpin.Flag(
		"opTimeout",
		"Timeout of operations per user (in seconds).",
	).Default("1").Uint()
	modifyTime = kingpin.Flag(
		"modifyTime",
		"How much to wait for user to modify its data (in seconds).",
	).Default("60").Uint()
	roundTime = kingpin.Flag(
		"roundTime",
		"Time for each round (in seconds).",
	).Default("10").Short('r').Uint()
	payTime = kingpin.Flag(
		"payTime",
		"Time to perform a payment (in minutes).",
	).Default("5").Uint()
	schedule = kingpin.Flag(
		"schedule",
		"Games schedule.",
	).Default("0 0 * * * *").Short('s').String()
	ticketPrice = kingpin.Flag(
		"ticketPrice",
		"Price of each ticket.",
	).Default("0.001").Short('p').Float64()
	dbPath = kingpin.Flag(
		"dbPath",
		"Path to the level db.",
	).Default("./db").Short('d').String()
	cashboxWalletPath = kingpin.Flag(
		"cashboxWalletPath",
		"Path to the \"cashbox\" wallet.",
	).Default("~/.electron-cash/wallets/cashbox_wallet").String()
	bankWalletPath = kingpin.Flag(
		"bankWalletPath",
		"Path to the \"bank\" wallet.",
	).Default("~/.electron-cash/wallets/bank_wallet").String()
	verbose = kingpin.Flag(
		"verbose",
		"Verbose logging mode.",
	).Short('v').Bool()
	token = kingpin.Arg(
		"token",
		"Bot's token.",
	).Required().String()
)

func main() {
	kingpin.Parse()
	opts := rps.NewOptions(
		*capacity,
		*timeout,
		*opTimeout,
		*modifyTime,
		*roundTime,
		*payTime,
		*schedule,
		*ticketPrice,
		*dbPath,
		*cashboxWalletPath,
		*bankWalletPath,
	)

	if *verbose {
		rps.LogsInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		rps.LogsInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

	users := rps.Users{}
	requests, stats, names := rps.LDBMap{}, rps.LDBMap{}, rps.LDBMap{}
	players := []int64{}
	leaderboard := []*rps.User{}

	crn := cron.New()
	crn.Start()
	bot := rps.New(*token, &opts, crn, &users, &requests, &stats, &names, players, &leaderboard)
	bot.Start()
}

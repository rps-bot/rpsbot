package rps

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// User structure.
type User struct {
	userID              int64
	subscribed          bool
	hasTicket           bool
	isPlayer            bool
	lastWonAmount       float64
	totalWonAmount      float64
	leaderboardPosition uint32
	playSequence        string
	name                string
	walletAddress       string
	lastTicketDate      time.Time
	joinDate            time.Time
	lock                *sync.RWMutex
}

// NewUser creates an object of User structure.
func NewUser(
	userID int64, name string,
	subscribed, hasTicket, isPlayer bool,
	leaderboardPosition uint32,
) User {
	var playSequence, walletAddress string
	var lastWonAmount, totalWonAmount float64
	var lastTicketDate, joinDate time.Time
	joinDate = time.Now()
	lock := sync.RWMutex{}

	u := User{userID, subscribed, hasTicket, isPlayer, lastWonAmount, totalWonAmount,
		leaderboardPosition, playSequence, name, walletAddress, lastTicketDate, joinDate, &lock}

	return u
}

// GetUserID performs non-blocking get of user's id.
func (u *User) GetUserID() int64 {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.userID
}

// GetSubscribed performs non-blocking get of user's subscription status.
func (u *User) GetSubscribed() bool {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.subscribed
}

// GetHasTicket performs non-blocking get of user's ticket status.
func (u *User) GetHasTicket() bool {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.hasTicket
}

// GetIsPlayer performs non-blocking get of user's player status.
func (u *User) GetIsPlayer() bool {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.isPlayer
}

// GetLastWonAmount performs non-blocking get of user's last won amount.
func (u *User) GetLastWonAmount() float64 {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.lastWonAmount
}

// GetTotalWonAmount performs non-blocking get of user's total won amount.
func (u *User) GetTotalWonAmount() float64 {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.totalWonAmount
}

// GetLeaderboardPosition performs non-blocking get of user's leaderboard position.
func (u *User) GetLeaderboardPosition() uint32 {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.leaderboardPosition
}

// GetPlaySequence performs non-blocking get of user's play sequence.
func (u *User) GetPlaySequence() string {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.playSequence
}

// GetName performs non-blocking get of user's name.
func (u *User) GetName() string {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.name
}

// GetWalletAddress performs non-blocking get of user's wallet address.
func (u *User) GetWalletAddress() string {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.walletAddress
}

// GetLastTicketDate performs non-blocking get of user's last ticket purchase date.
func (u *User) GetLastTicketDate() time.Time {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.lastTicketDate
}

// GetJoinDate performs non-blocking get of user's join date.
func (u *User) GetJoinDate() time.Time {
	(*u.lock).RLock()
	defer (*u.lock).RUnlock()
	return u.joinDate
}

// SetUserID performs non-blocking set of user's ID.
func (u *User) SetUserID(id int64) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.userID = id
}

// SetSubscribed performs non-blocking set of user's subscription status.
func (u *User) SetSubscribed(val bool) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.subscribed = val
}

// SetHasTicket performs non-blocking set of user's ticket status.
func (u *User) SetHasTicket(val bool) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.hasTicket = val
}

// SetIsPlayer performs non-blocking set of user's player status.
func (u *User) SetIsPlayer(val bool) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.isPlayer = val
}

// SetLastWonAmount performs non-blocking set of user's last won amount.
func (u *User) SetLastWonAmount(val float64) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.lastWonAmount = val
}

// SetTotalWonAmount performs non-blocking set of user's total won amount.
func (u *User) SetTotalWonAmount(val float64) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.totalWonAmount = val
}

// SetLeaderboardPosition performs non-blocking set of user's leaderboard position.
func (u *User) SetLeaderboardPosition(val uint32) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.leaderboardPosition = val
}

// SetPlaySequence performs non-blocking set of user's play sequence.
func (u *User) SetPlaySequence(val string) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.playSequence = val
}

// SetName performs non-blocking set of user's name.
func (u *User) SetName(val string) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.name = val
}

// SetWalletAddress performs non-blocking set of user's wallet address.
func (u *User) SetWalletAddress(val string) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.walletAddress = val
}

// SetLastTicketDate performs non-blocking set of user's last ticket purchase date.
func (u *User) SetLastTicketDate(date time.Time) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.lastTicketDate = date
}

// SetJoinDate performs non-blocking set of user's join date.
func (u *User) SetJoinDate(date time.Time) {
	(*u.lock).Lock()
	defer (*u.lock).Unlock()
	u.joinDate = date
}

// Serialize performs serialization of the User structure.
func (u *User) Serialize() []byte {
	return []byte(fmt.Sprintf("UserID: %d|Subscribed: %t|HasTicket: %t|IsPlayer: %t|"+
		"LastWonAmount: %f|TotalWonAmount: %f|LeaderboardPosition: %d|PlaySequence: %s|"+
		"Name: %s|WalletAddress: %s|LastTicketDate: %s|JoinDate: %s",
		u.userID, u.subscribed, u.hasTicket, u.isPlayer, u.lastWonAmount, u.totalWonAmount,
		u.leaderboardPosition, u.playSequence, u.name, u.walletAddress,
		u.lastTicketDate.Format(time.RFC1123), u.joinDate.Format(time.RFC1123)),
	)
}

// Deserialize performs deserialization of the User structure.
func Deserialize(data []byte) (User, error) {
	strData := string(data)
	d := strings.Split(strData, "|")

	strUserID := d[0][strings.Index(d[0], " ")+1:]
	userID, err := strconv.ParseInt(strUserID, 10, 64)
	if err != nil {
		return User{}, err
	}
	strSubscribed := d[1][strings.Index(d[1], " ")+1:]
	subscribed, err := strconv.ParseBool(strSubscribed)
	if err != nil {
		return User{}, err
	}
	strHasTicket := d[2][strings.Index(d[2], " ")+1:]
	hasTicket, err := strconv.ParseBool(strHasTicket)
	if err != nil {
		return User{}, err
	}
	strIsPlayer := d[3][strings.Index(d[3], " ")+1:]
	isPlayer, err := strconv.ParseBool(strIsPlayer)
	if err != nil {
		return User{}, err
	}
	strLastWonAmount := d[4][strings.Index(d[4], " ")+1:]
	lastWonAmount, err := strconv.ParseFloat(strLastWonAmount, 64)
	if err != nil {
		return User{}, err
	}
	strTotalWonAmount := d[5][strings.Index(d[5], " ")+1:]
	totalWonAmount, err := strconv.ParseFloat(strTotalWonAmount, 64)
	if err != nil {
		return User{}, err
	}
	strLeaderboardPosition := d[6][strings.Index(d[6], " ")+1:]
	_leaderboardPosition, err := strconv.ParseUint(strLeaderboardPosition, 10, 32)
	if err != nil {
		return User{}, err
	}
	leaderboardPosition := uint32(_leaderboardPosition)
	playSequence := d[7][strings.Index(d[7], " ")+1:]
	name := d[8][strings.Index(d[8], " ")+1:]
	walletAddress := d[9][strings.Index(d[9], " ")+1:]
	strLastTicketDate := d[10][strings.Index(d[10], " ")+1:]
	lastTicketDate, err := time.Parse(time.RFC1123, strLastTicketDate)
	if err != nil {
		return User{}, err
	}
	strJoinDate := d[11][strings.Index(d[11], " ")+1:]
	joinDate, err := time.Parse(time.RFC1123, strJoinDate)
	if err != nil {
		return User{}, err
	}
	lock := sync.RWMutex{}

	u := User{userID, subscribed, hasTicket, isPlayer, lastWonAmount, totalWonAmount,
		leaderboardPosition, playSequence, name, walletAddress, lastTicketDate, joinDate, &lock}

	return u, err
}

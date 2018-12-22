package rps

import (
	"os/exec"
	"regexp"
	"strings"
)

// ExecCMD executes commands in shell.
func ExecCMD(cmd string) ([]byte, error) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// ContainsInt64 checks if slice contains an element.
func ContainsInt64(e int64, s []int64) int {
	for i, el := range s {
		if el == e {
			return i
		}
	}
	return -1
}

// ContainsChar checks if string contains a character.
func ContainsChar(e rune, s string) int {
	for i, el := range s {
		if el == e {
			return i
		}
	}
	return -1
}

// CountChar counts number of entries of character in the string.
func CountChar(e rune, s string) uint {
	j := uint(0)
	for _, el := range s {
		if el == e {
			j++
		}
	}
	return j
}

// Reverse reverses a slice of *Users.
func Reverse(a *[]*User) {
	for i := 0; i < len(*a)/2; i++ {
		(*a)[i], (*a)[len(*a)-i-1] = (*a)[len(*a)-i-1], (*a)[i]
	}
}

// Max determines maximum of two values.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinFloat64 determines minumum of two values.
func MinFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat64 determines maximum of two values.
func MaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Min determines minumum of two values.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NameValidate validates username.
func NameValidate(name string) bool {
	if len(name) < 4 || len(name) > 16 {
		return false
	}
	re := regexp.MustCompile("[a-zA-Z0-9- ]+")
	if len(re.FindString(name)) != len(name) {
		return false
	}
	return true
}

// WalletValidate validates wallet address.
func WalletValidate(wallet string) bool {
	_wallet := ""

	idx := strings.Index(wallet, ":")
	if idx != -1 {
		_wallet = wallet[idx+1:]
	}

	if len(_wallet) != 42 {
		return false
	}

	return true
}

func mergeLeaderboardPosition(a []*User, b []*User) []*User {
	var r = make([]*User, len(a)+len(b))
	var i = 0
	var j = 0

	for i < len(a) && j < len(b) {

		if a[i].GetLeaderboardPosition() <= b[j].GetLeaderboardPosition() {
			r[i+j] = a[i]
			i++
		} else {
			r[i+j] = b[j]
			j++
		}

	}

	for i < len(a) {
		r[i+j] = a[i]
		i++
	}
	for j < len(b) {
		r[i+j] = b[j]
		j++
	}

	return r
}

// MergesortLeaderboardPosition performs merge sort by leaderboard position.
func MergesortLeaderboardPosition(items []*User) []*User {
	if len(items) < 2 {
		return items
	}

	var middle = len(items) / 2
	var a = MergesortLeaderboardPosition(items[:middle])
	var b = MergesortLeaderboardPosition(items[middle:])

	return mergeLeaderboardPosition(a, b)
}

func mergeLastTicketDate(a []*User, b []*User) []*User {
	var r = make([]*User, len(a)+len(b))
	var i = 0
	var j = 0

	for i < len(a) && j < len(b) {

		if a[i].GetLastTicketDate().Before(b[j].GetLastTicketDate()) {
			r[i+j] = a[i]
			i++
		} else {
			r[i+j] = b[j]
			j++
		}

	}

	for i < len(a) {
		r[i+j] = a[i]
		i++
	}
	for j < len(b) {
		r[i+j] = b[j]
		j++
	}

	return r
}

// MergesortLastTicketDate performs merge sort by last buy ticket date.
func MergesortLastTicketDate(items []*User) []*User {
	if len(items) < 2 {
		return items
	}

	var middle = len(items) / 2
	var a = MergesortLastTicketDate(items[:middle])
	var b = MergesortLastTicketDate(items[middle:])

	return mergeLastTicketDate(a, b)
}

func mergeJoinDate(a []*User, b []*User) []*User {
	var r = make([]*User, len(a)+len(b))
	var i = 0
	var j = 0

	for i < len(a) && j < len(b) {

		if a[i].GetJoinDate().Before(b[j].GetJoinDate()) {
			r[i+j] = a[i]
			i++
		} else {
			r[i+j] = b[j]
			j++
		}

	}

	for i < len(a) {
		r[i+j] = a[i]
		i++
	}
	for j < len(b) {
		r[i+j] = b[j]
		j++
	}

	return r
}

// MergesortJoinDate performs merge sort by join date.
func MergesortJoinDate(items []*User) []*User {
	if len(items) < 2 {
		return items
	}

	var middle = len(items) / 2
	var a = MergesortJoinDate(items[:middle])
	var b = MergesortJoinDate(items[middle:])

	return mergeJoinDate(a, b)
}

func mergeTotalWonAmount(a []*User, b []*User) []*User {
	var r = make([]*User, len(a)+len(b))
	var i = 0
	var j = 0

	for i < len(a) && j < len(b) {

		if a[i].GetTotalWonAmount() <= b[j].GetTotalWonAmount() {
			r[i+j] = a[i]
			i++
		} else {
			r[i+j] = b[j]
			j++
		}

	}

	for i < len(a) {
		r[i+j] = a[i]
		i++
	}
	for j < len(b) {
		r[i+j] = b[j]
		j++
	}

	return r
}

// MergesortTotalWonAmount performs merge sort by total won amount.
func MergesortTotalWonAmount(items []*User) []*User {
	if len(items) < 2 {
		return items
	}

	var middle = len(items) / 2
	var a = MergesortTotalWonAmount(items[:middle])
	var b = MergesortTotalWonAmount(items[middle:])

	return mergeTotalWonAmount(a, b)
}

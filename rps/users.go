package rps

// Users structure.
type Users struct {
	data UMap
}

// NewUsers creates an object of Users structure.
func NewUsers(name string, dbPath string) Users {
	data := NewUMap(name, dbPath)

	return Users{data}
}

// Apply performs an action on Users structure object.
func (u Users) Apply(uid int64, fn func(u *User) error) error {
	user := u.data.Get(uid)
	err := fn(user)
	if err != nil {
		return err
	}

	return u.data.Put(uid, user)
}

// Get an objecto from the vault.
func (u Users) Get(uid int64) *User {
	return u.data.Get(uid)
}

// Put an object to the vault.
func (u Users) Put(uid int64, user *User) error {
	return u.data.Put(uid, user)
}

// Delete an object from the vault.
func (u Users) Delete(uid int64) error {
	return u.data.Delete(uid)
}

// Exist checks if an object in the vault.
func (u Users) Exist(uid int64) bool {
	return u.data.Exist(uid)
}

// BatchPut puts an object into the vault by batching (leveldb).
func (u Users) BatchPut(uid int64, user *User) {
	u.data.BatchPut(uid, user)
}

// BatchDelete deletes an object from the vault by batching (leveldb).
func (u Users) BatchDelete(uid int64) {
	u.data.BatchDelete(uid)
}

// BatchWrite performs write of the batch to the database.
func (u Users) BatchWrite() error {
	return u.data.BatchWrite()
}

// Iterate brings possibility to iterate over the vault.
func (u Users) Iterate() map[int64]*User {
	return u.data.Iterate()
}

// Len returns length of the vault
func (u Users) Len() int {
	return u.data.Len()
}

// Close closes the database.
func (u Users) Close() {
	u.data.Close()
}

// FormTotalWonAmountList forms list of *User sorted by total won amount.
func (u Users) FormTotalWonAmountList() []*User {
	lst := make([]*User, u.data.Len())
	i := 0
	for _, user := range u.data.Iterate() {
		lst[i] = user
		i++
	}

	lst = MergesortJoinDate(lst)
	Reverse(&lst)
	return MergesortTotalWonAmount(lst)
}

// FormLeaderboardPositionList forms list of *User sorted by position in the leaderboard.
func (u Users) FormLeaderboardPositionList() []*User {
	lst := make([]*User, u.data.Len())
	i := 0
	for _, user := range u.data.Iterate() {
		lst[i] = user
		i++
	}

	lst = MergesortJoinDate(lst)
	Reverse(&lst)
	return MergesortLeaderboardPosition(lst)
}

// FormLastTicketDateList forms list of *User sorted by last ticket purchase date.
func (u Users) FormLastTicketDateList() []*User {
	lst := make([]*User, u.data.Len())
	i := 0
	for _, user := range u.data.Iterate() {
		lst[i] = user
		i++
	}

	lst = MergesortLastTicketDate(lst)
	Reverse(&lst)
	return lst
}

// FormJoinDateList forms list of *User sorted by join date.
func (u Users) FormJoinDateList() []*User {
	lst := make([]*User, u.data.Len())
	i := 0
	for _, user := range u.data.Iterate() {
		lst[i] = user
		i++
	}

	return MergesortJoinDate(lst)
}

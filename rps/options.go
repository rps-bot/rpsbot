package rps

// Options structure.
type Options struct {
	capacity          uint
	timeout           uint
	opTimeout         uint
	modifyTime        uint
	roundTime         uint
	payTime           uint
	schedule          string
	ticketPrice       float64
	dbPath            string
	cashboxWalletPath string
	bankWalletPath    string
}

// NewOptions creates an object of NewOptions structure.
func NewOptions(
	capacity uint,
	timeout uint,
	opTimeout uint,
	modifyTime uint,
	roundTime uint,
	payTime uint,
	schedule string,
	ticketPrice float64,
	dbPath string,
	cashboxWalletPath string,
	bankWalletPath string,
) Options {
	return Options{
		capacity, timeout, opTimeout, modifyTime, roundTime,
		payTime, schedule, ticketPrice, dbPath, cashboxWalletPath, bankWalletPath,
	}
}

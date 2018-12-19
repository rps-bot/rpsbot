package rps

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// GetBalance returns current balance of specified wallet
func GetBalance(walletPath string) (float64, error) {
	var request map[string]json.RawMessage
	var _confirmed, _unconfirmed string

	_res, err := ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s getbalance",
		walletPath))
	if err != nil {
		if string(_res) == "false\n" {
			return 0, errors.New("payto return false")
		}
		return 0, err
	}

	err = json.Unmarshal(_res, &request)
	if err != nil {
		return 0, err
	}
	json.Unmarshal(request["confirmed"], &_confirmed)
	if _, ok := request["unconfirmed"]; ok {
		json.Unmarshal(request["unconfirmed"], &_unconfirmed)
	}
	confirmed, err := strconv.ParseFloat(string(_confirmed), 64)
	if err != nil {
		return 0, err
	}
	if _unconfirmed != "" {
		unconfirmed, err := strconv.ParseFloat(string(_unconfirmed), 64)
		if err != nil {
			return 0, err
		}
		return confirmed + unconfirmed, nil
	}

	return confirmed, nil
}

// GetRequest returns request's metadata
func GetRequest(requestID string, walletPath string) (map[string]json.RawMessage, error) {
	var request map[string]json.RawMessage

	res, err := ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s getrequest %s",
		walletPath, requestID))
	if err != nil {
		return map[string]json.RawMessage{}, err
	}

	if err := json.Unmarshal(res, &request); err != nil {
		return map[string]json.RawMessage{}, err
	}

	return request, nil
}

// PayToUser performs pay to the specified user
func PayToUser(user *User, amount float64, walletPath string) error {
	if user.GetWalletAddress() != "" {
		if err := PayTo(user.GetWalletAddress(), amount, walletPath); err != nil {
			return err
		}
	}

	return nil
}

// PayTo performs pay to specified address
func PayTo(dstAddress string, amount float64, walletPath string) error {
	var request map[string]json.RawMessage
	var hexID string
	var res []byte
	var err error

	if amount != -1 {
		res, err = ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s payto %s %f",
			walletPath, dstAddress, amount))
	} else {
		res, err = ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s payto %s !",
			walletPath, dstAddress))
	}
	if err != nil {
		if string(res) == "false\n" {
			return errors.New("payto return false")
		}
		return err
	}

	err = json.Unmarshal(res, &request)
	if err != nil {
		return err
	}
	json.Unmarshal(request["hex"], &hexID)

	res, err = ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s broadcast %s",
		walletPath, hexID))
	if err != nil {
		if string(res) == "false\n" {
			return errors.New("broadcast return false")
		}
		return err
	}

	return nil
}

// CreateRequest creates payment request
func CreateRequest(amount float64, walletPath string) (string, string, error) {
	var request map[string]json.RawMessage
	var address, url string

	res, err := ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s addrequest %f",
		walletPath, amount))
	if err != nil {
		return "", "", err
	}

	err = json.Unmarshal(res, &request)
	if err != nil {
		if string(res) == "false\n" {
			return "", "", errors.New("addrequest return false")
		}
		return "", "", err
	}

	if err := json.Unmarshal(request["address"], &address); err != nil {
		return "", "", err
	}
	if err := json.Unmarshal(request["URI"], &url); err != nil {
		return "", "", err
	}

	return address, url, nil
}

// RemoveRequest removes payment request
func RemoveRequest(requestID string, walletPath string) error {
	res, err := ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s rmrequest %s",
		walletPath, requestID))
	if err != nil {
		return err
	}
	if string(res) == "false\n" {
		return errors.New("rmrequest return false")
	}

	return nil
}

// ClearRequests removes all active requests
func ClearRequests(walletPath string) error {
	res, err := ExecCMD(fmt.Sprintf("electron-cash --testnet -w %s clearrequests",
		walletPath))
	if err != nil {
		return err
	}
	if string(res) == "false\n" {
		return errors.New("clearrequests return false")
	}

	return nil
}

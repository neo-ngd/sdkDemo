package main

import (
	"github.com/joeqian10/neo-gogogo/wallet"
	"math/big"
)

type AccountAndPay struct {
	Account *wallet.Account
	AssetId string
	TxId    string
	N       int
	Value   float64
}

type AccountAndPaySlice []AccountAndPay

func (us AccountAndPaySlice) Len() int {
	return len(us)
}

func (us AccountAndPaySlice) Less(i int, j int) bool {
	return us[i].Value < us[j].Value
}

func (us AccountAndPaySlice) Swap(i, j int) {
	us[i], us[j] = us[j], us[i]
}

func (us AccountAndPaySlice) RemoveAt(index int) []AccountAndPay {
	length := len(us)
	if index < 0 || index >= length {
		return us
	}
	tmp := append(us[:index], us[index+1:]...)
	return tmp
}

type Nep5AccountAndPay struct {
	Account *wallet.Account
	AssetId string
	Value   *big.Int
}

type Nep5AccountAndPaySlice []Nep5AccountAndPay

func (us Nep5AccountAndPaySlice) Len() int {
	return len(us)
}

func (us Nep5AccountAndPaySlice) Less(i int, j int) bool {
	return us[i].Value.Cmp(us[j].Value) < 0
}

func (us Nep5AccountAndPaySlice) Swap(i, j int) {
	us[i], us[j] = us[j], us[i]
}

func (us Nep5AccountAndPaySlice) RemoveAt(index int) []Nep5AccountAndPay {
	length := len(us)
	if index < 0 || index >= length {
		return us
	}
	tmp := append(us[:index], us[index+1:]...)
	return tmp
}

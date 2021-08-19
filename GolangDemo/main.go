package main

import (
	"fmt"
	"github.com/joeqian10/neo-gogogo/helper"
	"github.com/joeqian10/neo-gogogo/sc"
	"github.com/joeqian10/neo-gogogo/tx"
	"github.com/joeqian10/neo-gogogo/wallet"
	"log"
	"math/big"
	"sort"
)

var blackHoleAddress = "AJ36ZCpMhiHYMdMAUaP7i1i9pJz4jMdiQV"
var YourNep5GateValue, _ = big.NewInt(0).SetString("10", 10) // use big.Int for nep5 amount, 10 for nneo, 2000000000 for cgas
var ZERO = big.NewInt(0)

type DemoHelper struct {
	tb *tx.TransactionBuilder
	w  *wallet.Wallet
}

func NewDemoHelper(rpcUrl string, walletFilePath string) (*DemoHelper, error) {
	tb := tx.NewTransactionBuilder(rpcUrl)
	w, err := wallet.NewWalletFromFile(walletFilePath)
	if err != nil {
		return nil, err
	}
	return &DemoHelper{
		tb: tb,
		w:  w,
	}, nil
}

func NewDemoHelper2(rpcUrl string, accounts []*wallet.Account) (*DemoHelper, error) {
	tb := tx.NewTransactionBuilder(rpcUrl)
	w := wallet.NewWallet()
	for _, a := range accounts {
		w.AddAccount(a)
	}
	return &DemoHelper{
		tb: tb,
		w:  w,
	}, nil
}

func (this *DemoHelper) MigrateUtxo(assetId string, amount float64, n3Address string) error {
	// get AccountAndPay
	aaps, sum, err := this.getAccountAndPay(this.w.Accounts, assetId, amount)
	if err != nil {
		return err
	}

	accountMap := map[string]*wallet.Account{}

	inputs := make([]*tx.CoinReference, 0)
	outputs := make([]*tx.TransactionOutput, 0)

	// handle inputs
	for _, aap := range aaps {
		txHash, err := helper.UInt256FromString(aap.TxId)
		if err != nil {
			return err
		}
		cf := &tx.CoinReference{
			PrevHash:  txHash,
			PrevIndex: uint16(aap.N),
		}
		inputs = append(inputs, cf)

		if _, ok := accountMap[aap.Account.Address]; !ok {
			accountMap[aap.Account.Address] = aap.Account
		}
	}

	// handle outputs
	assetHash, err := helper.UInt256FromString(assetId)
	if err != nil {
		return err
	}
	fixed8 := helper.Fixed8FromFloat64(amount)
	toScriptHash, err := helper.AddressToScriptHash(blackHoleAddress)
	if err != nil {
		return err
	}

	//
	changeAddr, err := helper.AddressToScriptHash(this.w.Accounts[0].Address)
	if err != nil {
		return err
	}

	output := tx.NewTransactionOutput(assetHash, fixed8, toScriptHash)
	outputs = append(outputs, output)

	// check asset type
	if assetId == "0x"+tx.NeoTokenId {
		if amount >= 10 {
			if sum > amount {
				outputs = append(outputs, tx.NewTransactionOutput(assetHash, helper.Fixed8FromFloat64(sum-amount), changeAddr))
			}
		} else {
			gasAaps, gasSum, err := this.getAccountAndPay(this.w.Accounts, "0x"+tx.NeoTokenId, 1)
			if err != nil {
				return err
			}
			for _, ga := range gasAaps {
				txHash, err := helper.UInt256FromString(ga.TxId)
				if err != nil {
					return err
				}
				cf := &tx.CoinReference{
					PrevHash:  txHash,
					PrevIndex: uint16(ga.N),
				}
				inputs = append(inputs, cf)

				if _, ok := accountMap[ga.Account.Address]; !ok {
					accountMap[ga.Account.Address] = ga.Account
				}
			}
			if gasSum > 1 {
				gasAssetHash, err := helper.UInt256FromString("0x" + tx.GasTokenId)
				if err != nil {
					return err
				}
				outputs = append(outputs, tx.NewTransactionOutput(gasAssetHash, helper.Fixed8FromFloat64(sum-amount), changeAddr))
			}
		}
	} else if assetId == "0x"+tx.GasTokenId {
		if amount >= 20 {
			if sum > amount {
				outputs = append(outputs, tx.NewTransactionOutput(assetHash, helper.Fixed8FromFloat64(sum-amount), changeAddr))
			}
		} else {
			gasAaps, gasSum, err := this.getAccountAndPay(this.w.Accounts, "0x"+tx.NeoTokenId, amount+1)
			if err != nil {
				return err
			}
			inputs = make([]*tx.CoinReference, 0)
			accountMap := map[string]*wallet.Account{}
			for _, ga := range gasAaps {
				txHash, err := helper.UInt256FromString(ga.TxId)
				if err != nil {
					return err
				}
				cf := &tx.CoinReference{
					PrevHash:  txHash,
					PrevIndex: uint16(ga.N),
				}
				inputs = append(inputs, cf)
				if _, ok := accountMap[ga.Account.Address]; !ok {
					accountMap[ga.Account.Address] = ga.Account
				}
			}
			if gasSum > amount+1 {
				gasAssetHash, err := helper.UInt256FromString("0x" + tx.GasTokenId)
				if err != nil {
					return err
				}
				outputs = append(outputs, tx.NewTransactionOutput(gasAssetHash, helper.Fixed8FromFloat64(gasSum-(amount+1)), changeAddr))
			}
		}
	}

	// build transaction
	ctx := tx.NewContractTransaction()
	attr := &tx.TransactionAttribute{
		Usage: tx.NewTransactionAttributeUsageFromString("Remark14"),
		Data:  []byte(n3Address),
	}
	ctx.Attributes = []*tx.TransactionAttribute{attr}
	ctx.Inputs = inputs
	ctx.Outputs = outputs

	// sign
	for _, v := range accountMap {
		err := tx.AddSignature(ctx, v.KeyPair)
		if err != nil {
			return err
		}
	}

	response := this.tb.Client.SendRawTransaction(ctx.RawTransactionString())
	if response.HasError() {
		return fmt.Errorf(response.ErrorResponse.GetErrorInfo())
	}
	log.Println(ctx.HashString())
	return nil
}

func (this *DemoHelper) MigrateNep5(assetId string, amount *big.Int, n3Address string) error {
	// get AccountAndPay
	naaps, _, err := this.getNep5AccountAndPay(this.w.Accounts, assetId, amount)
	if err != nil {
		return err
	}

	accountMap := map[string]*wallet.Account{}
	assetHash, err := helper.UInt160FromString(assetId)
	t, _ := helper.AddressToScriptHash(blackHoleAddress)
	script := []byte{}
	for _, naap := range naaps {
		f, _ := helper.AddressToScriptHash(naap.Account.Address)

		sb := sc.NewScriptBuilder()
		cp1 := sc.ContractParameter{
			Type:  sc.Hash160,
			Value: f.Bytes(),
		}
		cp2 := sc.ContractParameter{
			Type:  sc.Hash160,
			Value: t.Bytes(),
		}
		cp3 := sc.ContractParameter{
			Type:  sc.Integer,
			Value: naap.Value,
		}
		sb.MakeInvocationScript(assetHash.Bytes(), "transfer", []sc.ContractParameter{cp1, cp2, cp3})
		script = append(script, sb.ToArray()...)

		if _, ok := accountMap[naap.Account.Address]; !ok {
			accountMap[naap.Account.Address] = naap.Account
		}
	}

	// get gas consumed
	checkWitnessHashes := []string{}
	for _, v := range accountMap {
		a, _ := helper.AddressToScriptHash(v.Address)
		checkWitnessHashes = append(checkWitnessHashes, a.String())
	}
	response := this.tb.Client.InvokeScript(helper.BytesToHex(script), "0000000000000000000000000000000000000000")
	if response.HasError() {
		return fmt.Errorf(response.GetErrorInfo())
	}
	if response.Result.State == "FAULT" { // use ScriptContainer in contract will cause engine fault
		return fmt.Errorf("engine fault")
	}
	gasConsumed, err := helper.Fixed8FromString(response.Result.GasConsumed)
	if err != nil {
		return err
	}
	gasConsumed = gasConsumed.Sub(helper.Fixed8FromInt64(10))
	if gasConsumed.LessThan(helper.Zero) || gasConsumed.Equal(helper.Zero) {
		gasConsumed = helper.Zero
	} else {
		gasConsumed = gasConsumed.Ceiling()
	}

	// build transaction
	itx := tx.NewInvocationTransaction(script)
	attr := &tx.TransactionAttribute{
		Usage: tx.NewTransactionAttributeUsageFromString("Remark14"),
		Data:  []byte(n3Address),
	}
	itx.Attributes = []*tx.TransactionAttribute{attr}
	itx.Gas = gasConsumed
	netFee := helper.Fixed8FromInt64(0)
	if amount.Cmp(YourNep5GateValue) < 0 {
		netFee = helper.Fixed8FromInt64(1)
	}
	fee := netFee.Add(itx.Gas)
	if itx.Size() > 1024 {
		fee = fee.Add(helper.Fixed8FromFloat64(0.001))
		fee = fee.Add(helper.Fixed8FromFloat64(float64(itx.Size()) * 0.00001))
	}

	// get gas
	changeAddr, _ := helper.AddressToScriptHash(this.w.Accounts[0].Address)
	gasAaps, gasSum, err := this.getAccountAndPay(this.w.Accounts, "0x"+tx.NeoTokenId, helper.Fixed8ToFloat64(fee))
	if err != nil {
		return err
	}
	inputs := make([]*tx.CoinReference, 0)
	for _, ga := range gasAaps {
		txHash, err := helper.UInt256FromString(ga.TxId)
		if err != nil {
			return err
		}
		cf := &tx.CoinReference{
			PrevHash:  txHash,
			PrevIndex: uint16(ga.N),
		}
		inputs = append(inputs, cf)
		if _, ok := accountMap[ga.Account.Address]; !ok {
			accountMap[ga.Account.Address] = ga.Account
		}
	}
	outputs := make([]*tx.TransactionOutput, 0)
	if gasSum > helper.Fixed8ToFloat64(fee) {
		gasAssetHash, err := helper.UInt256FromString("0x" + tx.GasTokenId)
		if err != nil {
			return err
		}
		outputs = append(outputs, tx.NewTransactionOutput(gasAssetHash, helper.Fixed8FromFloat64(gasSum).Sub(fee), changeAddr))
	}

	itx.Inputs = inputs
	itx.Outputs = outputs

	// sign
	for _, v := range accountMap {
		err := tx.AddSignature(itx, v.KeyPair)
		if err != nil {
			return err
		}
	}

	response2 := this.tb.Client.SendRawTransaction(itx.RawTransactionString())
	if response2.HasError() {
		return fmt.Errorf(response.ErrorResponse.GetErrorInfo())
	}
	log.Println(itx.HashString())
	return nil

}

func (this *DemoHelper) getAccountAndPay(accounts []*wallet.Account, assetId string, amount float64) ([]AccountAndPay, float64, error) {
	if amount == 0 {
		return nil, 0, nil
	}
	sum := float64(0)
	aaps := make([]AccountAndPay, 0, 32)
	for _, account := range accounts {
		response := this.tb.Client.GetUnspents(account.Address)
		if response.HasError() {
			return nil, 0, fmt.Errorf(response.GetErrorInfo())
		}
		balances := response.Result.Balances
		// check if there is enough balance of this asset in this account
		for _, balance := range balances {
			if "0x"+balance.AssetHash == assetId {
				sum += balance.Amount
				unspents := balance.Unspents
				for _, unspent := range unspents {
					aap := AccountAndPay{
						Account: account,
						AssetId: assetId,
						TxId:    unspent.Txid,
						N:       unspent.N,
						Value:   unspent.Value,
					}
					aaps = append(aaps, aap)
				}
			}
		}
	}

	if sum < amount {
		return nil, 0, fmt.Errorf("insufficient funds of " + assetId)
	}
	// sort in descending order
	sort.Sort(sort.Reverse(AccountAndPaySlice(aaps)))

	results := make([]AccountAndPay, 0)
	sumRe := float64(0)
	var i int = 0
	var a = amount
	for i < len(aaps) && aaps[i].Value >= a {
		a -= aaps[i].Value
		results = append(results, aaps[i])
		sumRe += aaps[i].Value
		i++
	}
	if a == 0 {
		return results, sumRe, nil
	}

	for i < len(aaps) && aaps[i].Value >= a {
		i++
	}
	results = append(results, aaps[i-1])
	sumRe += aaps[i-1].Value

	return results, sumRe, nil
}

// amount must multiply D
func (this *DemoHelper) getNep5AccountAndPay(accounts []*wallet.Account, assetId string, amount *big.Int) ([]Nep5AccountAndPay, *big.Int, error) {
	assetScriptHash, _ := helper.UInt160FromString(assetId)

	if amount.Cmp(ZERO) == 0 {
		return nil, ZERO, nil
	}
	sum := big.NewInt(0)
	naaps := make([]Nep5AccountAndPay, 0)
	for _, account := range accounts {
		addressScriptHash, _ := helper.AddressToScriptHash(account.Address)

		sb := sc.NewScriptBuilder()
		cp := sc.ContractParameter{
			Type:  sc.Hash160,
			Value:  addressScriptHash.Bytes(),
		}
		sb.MakeInvocationScript(assetScriptHash.Bytes(), "balanceOf", []sc.ContractParameter{cp})
		script := sb.ToArray()
		response := this.tb.Client.InvokeScript(helper.BytesToHex(script), helper.ZeroScriptHashString)
		if response.HasError() {
			return nil, ZERO, fmt.Errorf(response.GetErrorInfo())
		}
		if response.Result.State == "FAULT" {
			return nil, ZERO, fmt.Errorf("engine faulted")
		}
		if len(response.Result.Stack) == 0 {
			return nil, ZERO, fmt.Errorf("no stack result returned")
		}
		stack := response.Result.Stack[0]
		bytes := helper.HexToBytes(stack.Value.(string))
		balance := helper.BigIntFromNeoBytes(bytes)
		sum.Add(sum, balance)

		naap := Nep5AccountAndPay{
			Account: account,
			AssetId: assetId,
			Value:   balance,
		}

		naaps = append(naaps, naap)
	}

	if sum.Cmp(amount) < 0 {
		return nil, ZERO, fmt.Errorf("insufficient funds of " + assetId)
	}
	// sort in descending order
	sort.Sort(sort.Reverse(Nep5AccountAndPaySlice(naaps)))

	results := make([]Nep5AccountAndPay, 0)
	sumRe := big.NewInt(0)
	var i int = 0
	var a = amount
	for i < len(naaps) && naaps[i].Value.Cmp(a) >= 0 {
		a.Sub(a, naaps[i].Value)
		results = append(results, naaps[i])
		sumRe.Add(sumRe, naaps[i].Value)
		i++
	}
	if a.Cmp(ZERO) == 0 {
		return results, sumRe, nil
	}

	for i < len(naaps) && naaps[i].Value.Cmp(a) >= 0 {
		i++
	}
	results = append(results, naaps[i-1])
	sumRe.Add(sumRe, naaps[i-1].Value)

	return results, sumRe, nil
}

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

//Should do a docker-compose down after this to make sure cache and db get nuked
func TestMain(m *testing.M) {
	createAndOpenDB()
	initCache()
	createUser(userId)
	stock := &Stock{Symbol: symbol, Price: quotePrice, Hash: hash, TimeStamp: time.Now()}
	setCache(stock.Symbol, stock)
	os.Exit(m.Run())
}

func TestQuote(t *testing.T) {
	resp, err := s.Quote(context.Background(), genGrpcRequest("QUOTE"))
	if err != nil {
		t.Error("TestQuote got unexpected error: ", err)
	}
	if !strings.Contains(resp.Message, hash) {
		t.Errorf("TestQuote %v didnt contain %v", resp.Message, hash)
	}
}

func TestAdd(t *testing.T) {
	resp, err := s.Add(context.Background(), genGrpcRequest("ADD"))
	exp := &User{Balance: amount, Id: userId, BuyStack: []*Buy{}, SellStack: []*Sell{}}
	if err != nil {
		t.Error("TestAdd got unexpected error: ", err)
	}
	if resp.Message != toString(exp) {
		t.Errorf("TestAdd got %v wanted %v", resp.Message, toString(exp))
	}
	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != amount {
		t.Errorf("TestAdd expected balance of %v got %v", amount, balance)
	}
}

func TestBuy(t *testing.T) {
	s.Buy(context.Background(), genGrpcRequest("BUY"))
	user, _ := getCacheUser(userId)
	if len(user.BuyStack) == 0 {
		t.Errorf("TestBuy expected Buy stack to have length of %d but had %d", 1, 0)
	}
	buy := user.BuyStack[0]
	if buy.Price != quotePrice {
		t.Errorf("TestBuy expected quote price of %v got %v", quotePrice, buy.Price)
	}
	if buy.StockSymbol != symbol {
		t.Errorf("TestBuy expected symbol of %v got %v", symbol, buy.StockSymbol)
	}
	if buy.StockBoughtAmount != int(amount/quotePrice) {
		t.Errorf("TestBuy expected stockBought amount of %v got %v", amount/quotePrice, buy.StockBoughtAmount)
	}
	if buy.ActualCashAmount != float32(amount/quotePrice)*quotePrice {
		t.Errorf("TestBuy expected actual cash amount amount of %v got %v", float32(amount/quotePrice)*quotePrice, buy.ActualCashAmount)
	}
	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != 0 {
		t.Errorf("TestBuy expected balance of %v got %v", 0, balance)
	}
}

func TestCommitBuy(t *testing.T) {
	s.CommitBuy(context.Background(), genGrpcRequest("COMMIT_BUY"))
	user, _ := getCacheUser(userId)
	if len(user.BuyStack) == 1 {
		t.Errorf("TestCommitBuy expected Buy stack to have length of %d but had %d", 0, 1)
	}
	var stockAmount int
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != int(amount/quotePrice) {
		t.Errorf("TestCommitBuy expected stock amount of %v got %v", int(amount/quotePrice), stockAmount)
	}

	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != 0 {
		t.Errorf("TestCommitBuy expected balance of %v got %v", 0, balance)
	}
}

func TestSell(t *testing.T) {
	s.Sell(context.Background(), genGrpcRequest("SELL"))
	user, _ := getCacheUser(userId)
	if len(user.SellStack) == 0 {
		t.Errorf("TestSell expected Sell stack to have length of %d but had %d", 1, 0)
	}
	sell := user.SellStack[0]
	if sell.Price != quotePrice {
		t.Errorf("TestSell expected quote price of %v got %v", quotePrice, sell.Price)
	}
	if sell.StockSymbol != symbol {
		t.Errorf("TestSell expected symbol of %v got %v", symbol, sell.StockSymbol)
	}
	if sell.StockSoldAmount != int(amount/quotePrice) {
		t.Errorf("TestSell expected stockSold amount of %v got %v", amount/quotePrice, sell.StockSoldAmount)
	}
	if sell.ActualCashAmount != float32(amount/quotePrice)*quotePrice {
		t.Errorf("TestSell expected actual cash amount amount of %v got %v", float32(amount/quotePrice)*quotePrice, sell.ActualCashAmount)
	}

	var stockAmount int
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != 0 {
		t.Errorf("TestSell expected stock amount of %v got %v", 0, stockAmount)
	}
}

func TestCommitSell(t *testing.T) {
	s.CommitSell(context.Background(), genGrpcRequest("COMMIT_SELL"))
	user, _ := getCacheUser(userId)
	if len(user.BuyStack) == 1 {
		t.Errorf("TestCommitSell expected Sell stack to have length of %d but had %d", 0, 1)
	}
	var stockAmount int
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != 0 {
		t.Errorf("TestCommitSell expected stock amount of %v got %v", 0, stockAmount)
	}

	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != amount {
		t.Errorf("TestCommitSell expected balance of %v got %v", amount, balance)
	}
}

func TestCancelBuy(t *testing.T) {
	s.Buy(context.Background(), genGrpcRequest("BUY"))
	user, _ := getCacheUser(userId)
	if len(user.BuyStack) != 1 {
		t.Errorf("TestCancelBuy expected Buy stack to have length of %d but had %d", 1, len(user.BuyStack))
	}
	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != 0 {
		t.Errorf("TestCancelBuy expected balance of %v got %v", 0, balance)
	}
	s.CancelBuy(context.Background(), genGrpcRequest("CANCEL_BUY"))
	user, _ = getCacheUser(userId)
	if len(user.BuyStack) != 0 {
		t.Errorf("TestCancelBuy expected Buy stack to have length of %d but had %d", 0, len(user.BuyStack))
	}
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != amount {
		t.Errorf("TestCancelBuy expected balance of %v got %v", amount, balance)
	}
}

func TestCancelSell(t *testing.T) {
	s.Buy(context.Background(), genGrpcRequest("BUY"))
	s.CommitBuy(context.Background(), genGrpcRequest("COMMIT_BUY"))
	s.Sell(context.Background(), genGrpcRequest("SELL"))
	user, _ := getCacheUser(userId)
	if len(user.SellStack) == 0 {
		t.Errorf("TestCancelSell expected Sell stack to have length of %d but had %d", 1, 0)
	}
	var stockAmount int
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != 0 {
		t.Errorf("TestCancelSell expected stock amount of %v got %v", 0, stockAmount)
	}

	s.CancelSell(context.Background(), genGrpcRequest("CANCEL_SELL"))
	user, _ = getCacheUser(userId)
	if len(user.SellStack) != 0 {
		t.Errorf("TestCancelSell expected Sell stack to have length of %d but had %d", 0, len(user.SellStack))
	}
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != int(amount/quotePrice) {
		t.Errorf("TestCancelSell expected stock amount of %v got %v", int(amount/quotePrice), stockAmount)
	}
	s.Sell(context.Background(), genGrpcRequest("SELL"))
	s.CommitSell(context.Background(), genGrpcRequest("COMMIT_SELL"))
}

func TestSetBuyAmount(t *testing.T) {
	s.SetBuyAmount(context.Background(), genGrpcRequest("SET_BUY_AMOUNT"))
	var buyId int
	var cashAmount float32
	db.QueryRow("SELECT Id,IntendedCashAmount from Buy where Id=(Select max(Id) from Buy)").Scan(&buyId, &cashAmount)
	if cashAmount != amount {
		t.Errorf("TestSetBuyAmount expected intendedCashAmount to be %f but was %f", amount, cashAmount)
	}

	var active bool
	db.QueryRow("SELECT Active from Buy_Trigger where BuyId=? and UserId=?", buyId, userId).Scan(&active)
	if active {
		t.Errorf("TestSetBuyAmount expected trigger to be false but was true")
	}
}

func TestSetBuyTrigger(t *testing.T) {
	var buyId int
	var cashAmount float32
	var price float32
	var stockBoughtAmount int
	var active bool
	buyPrice := float32(3.0)

	req := genGrpcRequest("SET_BUY_TRIGGER")
	req.Amount = buyPrice
	s.SetBuyTrigger(context.Background(), req)
	db.QueryRow("SELECT Id,IntendedCashAmount,Price,StockBoughtAmount from Buy where Id=(Select max(Id) from Buy)").Scan(&buyId, &cashAmount, &price, &stockBoughtAmount)
	if cashAmount != amount {
		t.Errorf("TestSetBuyTrigger expected intendedCashAmount to be %f but was %f", amount, cashAmount)
	}
	if price != buyPrice {
		t.Errorf("TestSetBuyTrigger expected price to be %f but was %f", buyPrice, price)
	}
	if stockBoughtAmount != int(amount/buyPrice) {
		t.Errorf("TestSetBuyTrigger expected stockBoughtAmount to be %d but was %d", int(amount/buyPrice), stockBoughtAmount)
	}
	db.QueryRow("SELECT Active from Buy_Trigger where BuyId=? and UserId=?", buyId, userId).Scan(&active)
	if !active {
		t.Errorf("TestSetBuyTrigger expected trigger to be true but was false")
	}
}

func TestCancelSetBuy(t *testing.T) {
	s.CancelSetBuy(context.Background(), genGrpcRequest("CANCEL_SET_BUY"))
	_, err := db.Query("SELECT * from Buy_Trigger where UserId=?", userId)
	if err != nil {
		t.Error("TestCancelSetBuy Expected no buy trigger but one was returned")
	}
	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != amount {
		t.Errorf("TestCancelSetBuy expected balance of %v got %v", 0, amount)
	}
}

func TestSetSellAmount(t *testing.T) {
	s.Buy(context.Background(), genGrpcRequest("BUY"))
	s.CommitBuy(context.Background(), genGrpcRequest("COMMIT_BUY"))
	s.SetSellAmount(context.Background(), genGrpcRequest("SET_SELL_AMOUNT"))
	var sellId int
	var cashAmount float32
	db.QueryRow("SELECT Id,IntendedCashAmount from Sell where Id=(Select max(Id) from Sell)").Scan(&sellId, &cashAmount)
	if cashAmount != amount {
		t.Errorf("TestSetSellAmount expected intendedCashAmount to be %f but was %f", amount, cashAmount)
	}

	var active bool
	db.QueryRow("SELECT Active from Sell_Trigger where SellId=? and UserId=?", sellId, userId).Scan(&active)
	if active {
		t.Errorf("TestSetSellAmount expected trigger to be false but was true")
	}
}

func TestSetSellTrigger(t *testing.T) {
	var sellId int
	var cashAmount float32
	var price float32
	var stockSoldAmount int
	var active bool
	sellPrice := float32(6.0)

	req := genGrpcRequest("SET_SELL_TRIGGER")
	req.Amount = sellPrice
	s.SetSellTrigger(context.Background(), req)
	db.QueryRow("SELECT Id,IntendedCashAmount,Price,StockSoldAmount from Sell where Id=(Select max(Id) from Sell)").Scan(&sellId, &cashAmount, &price, &stockSoldAmount)
	if cashAmount != amount {
		t.Errorf("TestSetSellTrigger expected intendedCashAmount to be %f but was %f", amount, cashAmount)
	}
	if price != sellPrice {
		t.Errorf("TestSetSellTrigger expected price to be %f but was %f", sellPrice, price)
	}
	if stockSoldAmount != int(amount/sellPrice) {
		t.Errorf("TestSetSellTrigger expected stockSoldAmount to be %d but was %d", int(amount/sellPrice), stockSoldAmount)
	}
	db.QueryRow("SELECT Active from Sell_Trigger where SellId=? and UserId=?", sellId, userId).Scan(&active)
	if !active {
		t.Errorf("TestSetSellTrigger expected trigger to be true but was false")
	}
}

func TestCancelSetSell(t *testing.T) {
	s.CancelSetSell(context.Background(), genGrpcRequest("CANCEL_SET_SELL"))
	_, err := db.Query("SELECT * from Sell_Trigger where UserId=?", userId)
	if err != nil {
		t.Error("TestCancelSetSell Expected no sell trigger but one was returned")
	}
	var stockAmount int
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != int(amount/quotePrice) {
		t.Errorf("TestCancelSetSell expected stock amount of %v got %v", int(amount/quotePrice), stockAmount)
	}
}

func TestCheckSellTrigger(t *testing.T) {
	sellPrice := float32(4.0)
	s.SetSellAmount(context.Background(), genGrpcRequest("SET_SELL_AMOUNT"))
	req := genGrpcRequest("SET_SELL_TRIGGER")
	req.Amount = sellPrice
	s.SetSellTrigger(context.Background(), req)
	checkSellTriggers()

	var balance float32
	db.QueryRow("SELECT Balance from User where Id=?", userId).Scan(&balance)
	if balance != amount {
		t.Errorf("TestCheckSellTrigger expected balance of %v got %v", amount, balance)
	}
	_, err := db.Query("SELECT * from Buy_Trigger where UserId=?", userId)
	if err != nil {
		t.Error("TestCheckSellTrigger Expected no trigger but one was returned")
	}
}

func TestCheckBuyTrigger(t *testing.T) {
	buyPrice := float32(6.0)
	var stockAmount int
	s.SetBuyAmount(context.Background(), genGrpcRequest("SET_BUY_AMOUNT"))
	req := genGrpcRequest("SET_BUY_TRIGGER")
	req.Amount = buyPrice
	s.SetBuyTrigger(context.Background(), req)
	checkBuyTriggers()
	db.QueryRow("SELECT Amount from User_Stock where UserId=? and StockSymbol=?", userId, symbol).Scan(&stockAmount)
	if stockAmount != int(amount/quotePrice) {
		t.Errorf("TestCheckBuyTrigger expected stock amount of %v got %v", int(amount/buyPrice), stockAmount)
	}
	_, err := db.Query("SELECT * from Buy_Trigger where UserId=?", userId)
	if err != nil {
		t.Error("TestCheckBuyTrigger Expected no trigger but one was returned")
	}
}
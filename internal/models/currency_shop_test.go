package models

import "testing"

func TestCurrencyTotalCopper(t *testing.T) {
	c := Currency{Copper: 5, Silver: 2, Gold: 1} // 5 + 20 + 100
	if got := c.TotalCopper(); got != 125 {
		t.Fatalf("expected 125, got %d", got)
	}
}

func TestCurrencyFromCopper(t *testing.T) {
	c := CurrencyFromCopper(1234) // 1 pp, 2 gp, 3 sp, 4 cp
	if c.Platinum != 1 || c.Gold != 2 || c.Silver != 3 || c.Copper != 4 {
		t.Fatalf("unexpected distribution: %+v", c)
	}
}

func TestCurrencySpendCoins(t *testing.T) {
	c := Currency{Gold: 10}
	if err := c.SpendCoins(Currency{Gold: 3, Silver: 5}); err != nil { // cost 350 cp
		t.Fatalf("spend failed: %v", err)
	}
	if got := c.TotalCopper(); got != 650 {
		t.Errorf("expected 650 cp left, got %d", got)
	}
	if c.Gold != 6 || c.Silver != 5 {
		t.Errorf("unexpected remainder: %+v", c)
	}
}

func TestCurrencySpendCoinsInsufficient(t *testing.T) {
	c := Currency{Gold: 1}
	if err := c.SpendCoins(Currency{Gold: 2}); err != ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
	if c.Gold != 1 {
		t.Errorf("pouch should be unchanged on failure: %+v", c)
	}
}

func TestCurrencyAddValue(t *testing.T) {
	c := Currency{Gold: 1}
	c.AddValue(Currency{Silver: 5}) // +50 cp -> 150 cp = 1 gp 5 sp
	if c.Gold != 1 || c.Silver != 5 {
		t.Errorf("unexpected: %+v", c)
	}
}

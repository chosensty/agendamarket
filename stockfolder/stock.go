package stock

type Stock struct {
	name         string
	value        float64
	sensitivity  float64
	stockID      int
	total_shares int
}

func NewStock(name string, price float64, sensitivity float64, stockID int) *Stock {
	s := Stock{name: name, value: price, sensitivity: sensitivity, stockID: stockID}
	return &s
}
func StockName(s *Stock) string {
	return s.name
}

func StockPurchase(s *Stock, shares float64) int {
	s.value *= 1 + getStockDifference(shares, s)
	return 1
}

func StockSale(s *Stock, shares float64) int {
	s.value *= 1 - getStockDifference(shares, s)
	return 1
}

func StockValue(s *Stock) float64 {
	return s.value
}

func getStockDifference(share_num float64, s *Stock) float64 {
	return 0.001 * share_num
}

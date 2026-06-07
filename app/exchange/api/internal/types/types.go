package types

type ExchangeReq struct {
	IP          string  `json:"ip,optional" form:"ip,optional"`
	Symbol      string  `json:"symbol,optional" form:"symbol,optional"`
	PageNo      int64   `json:"pageNo,optional" form:"pageNo,optional"`
	PageSize    int64   `json:"pageSize,optional" form:"pageSize,optional"`
	Price       float64 `json:"price,optional" form:"price,optional"`
	Amount      float64 `json:"amount,optional" form:"amount,optional"`
	Direction   string  `json:"direction,optional" form:"direction,optional"`
	Type        string  `json:"type,optional" form:"type,optional"`
	UseDiscount float64 `json:"useDiscount,optional" form:"useDiscount,optional"`
}

func (r *ExchangeReq) OrderValid() bool {
	return r.Direction != "" && r.Type != ""
}

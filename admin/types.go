package admin

type updateSwapPairRequest struct {
	BEP20Addr  string `json:"bep20_addr"`
	Available  bool   `json:"available"`
	LowerBound string `json:"lower_bound"`
	UpperBound string `json:"upper_bound"`
	IconUrl    string `json:"icon_url"`
}

type withdrawTokenRequest struct {
	Chain     string `json:"chain"`
	TokenAddr string `json:"token_addr"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

type withdrawTokenResponse struct {
	TxHash string `json:"tx_hash"`
	ErrMsg string `json:"err_msg"`
}

type retryFailedSwapsRequest struct {
	SwapIDList []int64 `json:"swap_id_list"`
}

type retryFailedSwapsResponse struct {
	SwapIDList []int64 `json:"swap_id_list"`
	ErrMsg string `json:"err_msg"`
}

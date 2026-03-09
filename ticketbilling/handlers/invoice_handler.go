package handlers

type InvoiceRequest struct {
	F   string  `json:"f"`
	T   float64 `json:"t"`
	RFC string  `json:"rfc"`
	E   string  `json:"e"`
}

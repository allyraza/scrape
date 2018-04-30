package souqr

// Attributes represents product attrs in resp
type Attributes struct {
	EAN           string `json:"Alternative_EANs"`
	Brand         string `json:"Brand"`
	Size          string `json:"Size"`
	TargetedGroup string `json:"Targeted_Group"`
}

// SouqProduct represents a product in resp
type SouqProduct struct {
	ID             string     `json:"id"`
	ItemID         int        `json:"id_item"`
	Name           string     `json:"name"`
	Price          float32    `json:"price"`
	Discount       float32    `json:"discount"`
	Currency       string     `json:"currencyCode"`
	Quantity       int        `json:"quantity"`
	Brand          string     `json:"brand"`
	Category       string     `json:"category"`
	Variant        string     `json:"variant"`
	Attributes     Attributes `json:"attributes"`
	ParentCategory string     `json:"super_category"`
}

// PageData represents a page data map in resp
type PageData struct {
	ItemID     int         `json:"ItemIDs"`
	Category   string      `json:"channel_name,eomitempty"`
	Reviews    int         `json:"item_reviews"`
	Name       string      `json:"item_title"`
	PriceRange string      `json:"price_ranges"`
	EAN        string      `json:"s_ean"`
	SoldOut    string      `json:"sold_out"`
	Rating     int         `json:"s_item_rating_total"`
	AvgRating  string      `json:"s_item_rating_avg"`
	Product    SouqProduct `json:"product"`
}

// Souq respresents data collected from resp
type Souq struct {
	Name        string   `json:"name"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Language    string   `json:"s_language"`
	Country     string   `json:"s_country"`
	GTIN        string   `json:"gtin13"`
	Color       string   `json:"color"`
	PageData    PageData `json:"Page_Data"`
}

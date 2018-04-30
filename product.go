package souqr

import (
	"encoding/xml"
	"io"
	"log"
)

// Product represents the single product
type Product struct {
	ID             int      `xml:"product_id",json:"product_id"`
	Name           string   `xml:"title",json:"title"`
	FamilyCode     string   `xml:"family_code,omitempty",json:"family_code,omitempty"`
	CategoryPath   string   `xml:"category_path",json:"category_path"`
	Brand          string   `xml:"brand,omitempty",json:"brand,omitempty"`
	Deeplink       string   `xml:"deeplink",json:"deeplink"`
	Description    string   `xml:"description",json:"description"`
	Image          string   `xml:"image",json:"image"`
	Color          string   `xml:"color,omitempty",json:"color,omitempty"`
	Material       string   `xml:"material,omitempty",json:"material,omitempty"`
	Gender         string   `xml:"gender,omitempty",json:"gender,omitempty"`
	Size           string   `xml:"size,omitempty",json:"size,omitempty"`
	EAN            string   `xml:"ean,omitempty",json:"ean,omitempty"`
	Quantity       int      `xml:"stock_amount",json:"stock_amount"`
	InStock        int      `xml:"stock_status",json:"stock_status"`
	Price          float32  `xml:"price",json:"price"`
	PriceExVat     float32  `xml:"price_ex_vat",json:"price_ex_vat"`
	PriceFrom      float32  `xml:"price_from",json:"price_from"`
	FloorPrice     float32  `xml:"floor_price",json:"floor_price"`
	MaxCPO         float32  `xml:"max_cpo,omitempty",json:"max_cpo,omitempty"`
	AgeGroup       string   `xml:"age_group",json:"age_group"`
	DeliveryCost   float32  `xml:"delivery_cost",json:"delivery_cost"`
	DeliveryPeriod string   `xml:"delivery_period",json:"delivery_period"`
	XMLName        xml.Name `xml:"product",json:"product"`
}

func (p *Product) Write(w io.Writer) (int, error) {
	xmlStr, err := xml.MarshalIndent(p, "", "    ")
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	return w.Write(xmlStr)
}

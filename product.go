package souqr

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
)

var seq int

// Variant represents the product variant
type Variant struct {
	XMLName        xml.Name `xml:"variant"`
	ID             string   `xml:"variant_id",json:"variant_id"`
	Description    string   `xml:"description",json:"description"`
	Price          string   `xml:"price",json:"price"`
	Color          string   `xml:"color",json:"color"`
	EAN            string   `xml:"ean",json:"ean"`
	Size           string   `xml:"size",json:"size"`
	DeliveryPeriod string   `xml:"deliver_period",json:"deliver_period"`
}

// Product represents the single product
type Product struct {
	XMLName      xml.Name   `xml:"product",json:"product"`
	ID           int        `xml:"product_id",json:"product_id"`
	Name         string     `xml:"title",json:"title"`
	FamilyCode   string     `xml:"family_code",json:"family_code"`
	CategoryPath string     `xml:"category_path",json:"category_path"`
	Brand        string     `xml:"brand",json:"brand"`
	Deeplink     string     `xml:"deeplink",json:"deeplink"`
	Variants     []*Variant `xml:"variants>variant",json:"variants"`
}

// ProductList represents a list of products
type ProductList struct {
	XMLName  xml.Name   `xml:"products"`
	Products []*Product `json:"products"`
}

func generateID() string {
	seq = seq + 1
	return fmt.Sprintf("v%v", seq)
}

// AddVariant adds a variant to product
func (p *Product) AddVariant(v *Variant) {
	v.ID = generateID()
	p.Variants = append(p.Variants, v)
}

// AddProduct adds a given product to product list
func (pl *ProductList) AddProduct(p *Product) {
	pl.Products = append(pl.Products, p)
}

func (pl *ProductList) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&pl)
	return buf.String()
}

func (pl *ProductList) Write(w io.Writer) (int, error) {
	xmlStr, err := xml.MarshalIndent(pl, "", "    ")
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	return w.Write(xmlStr)
}

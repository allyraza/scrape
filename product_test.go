package souqr

import (
	"fmt"
	"os"
	"testing"
)

func TestProductToXML(t *testing.T) {
	v := &Variant{ID: 1, Description: "Red tshirt description", Color: "Red"}
	v2 := &Variant{ID: 2, Description: "Blue tshirt description", Color: "Blue"}

	p := &Product{ID: 1, Name: "Red Tshirt", CategoryPath: "Men > Clothing > Tshirts"}
	p2 := &Product{ID: 2, Name: "Blue Tshirt", CategoryPath: "Men > Clothing > Tshirts"}
	p.AddVariant(v)
	p2.AddVariant(v)
	p2.AddVariant(v2)

	pl := &ProductList{}
	pl.AddProduct(p)
	pl.AddProduct(p2)
	pl.Write(os.Stdout)

	fmt.Println("")
}

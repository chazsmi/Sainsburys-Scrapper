package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestToJson(t *testing.T) {
	// Mock a product struct
	product := Product{
		Title:       "This is a test",
		Size:        "10 size",
		UnitPrice:   1.00,
		Description: "this is a test description",
	}

	products := Products{Results: []Product{product}}

	// Create buffer to pass in as we are going to want to check the ouput
	var out bytes.Buffer
	ToJson(products, &out)

	// Check the out is Json
	var results Products
	if err := json.Unmarshal(out.Bytes(), &results); err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestMakeRequest(t *testing.T) {
	// Make a request to google
	bow := MakeRequest("https://www.google.co.uk")
	title := bow.Find("title").Text()
	if title != "Google" {
		t.Fail()
	}
}

func TestProcessPage(t *testing.T) {
	// Sample of a product page
	url := "http://www.sainsburys.co.uk/shop/gb/groceries/ripe---ready/sainsburys-avocado-xl-pinkerton-loose-300g"
	chProducts := make(chan Product)

	// Set the method off as a Go routine
	go ProcessPage(url, chProducts)

	for {
		// Wait for a response
		result := <-chProducts
		if result.Title != "Sainsbury's Avocado Ripe & Ready XL Loose 300g" {
			t.Fail()
		}
		return
	}

}

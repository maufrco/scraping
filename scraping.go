package scp

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//Anuncio exportar dados do scrap
type Anuncio struct {
	Quantidade int
	Vendido    int
	Permalink  string
	Titulo     string
}

//Result retorno do ML
type Result struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
	Title  string `json:"title"`
	Seller struct {
		ID                int           `json:"id"`
		Permalink         interface{}   `json:"permalink"`
		PowerSellerStatus interface{}   `json:"power_seller_status"`
		CarDealer         bool          `json:"car_dealer"`
		RealEstateAgency  bool          `json:"real_estate_agency"`
		Tags              []interface{} `json:"tags"`
	} `json:"seller"`
	Price              float64   `json:"price"`
	CurrencyID         string    `json:"currency_id"`
	AvailableQuantity  int       `json:"available_quantity"`
	SoldQuantity       int       `json:"sold_quantity"`
	BuyingMode         string    `json:"buying_mode"`
	ListingTypeID      string    `json:"listing_type_id"`
	StopTime           time.Time `json:"stop_time"`
	Condition          string    `json:"condition"`
	Permalink          string    `json:"permalink"`
	Thumbnail          string    `json:"thumbnail"`
	AcceptsMercadopago bool      `json:"accepts_mercadopago"`
	Installments       struct {
		Quantity   int     `json:"quantity"`
		Amount     float64 `json:"amount"`
		Rate       float64 `json:"rate"`
		CurrencyID string  `json:"currency_id"`
	} `json:"installments"`
	Address struct {
		StateID   string `json:"state_id"`
		StateName string `json:"state_name"`
		CityID    string `json:"city_id"`
		CityName  string `json:"city_name"`
	} `json:"address"`
	Shipping struct {
		FreeShipping bool     `json:"free_shipping"`
		Mode         string   `json:"mode"`
		Tags         []string `json:"tags"`
		LogisticType string   `json:"logistic_type"`
		StorePickUp  bool     `json:"store_pick_up"`
	} `json:"shipping"`
	SellerAddress struct {
		ID          string `json:"id"`
		Comment     string `json:"comment"`
		AddressLine string `json:"address_line"`
		ZipCode     string `json:"zip_code"`
		Country     struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"country"`
		State struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"state"`
		City struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"city"`
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"seller_address"`
	Attributes []struct {
		AttributeGroupID string `json:"attribute_group_id"`
		ID               string `json:"id"`
		ValueID          string `json:"value_id"`
		ValueName        string `json:"value_name"`
		Values           []struct {
			Struct interface{} `json:"struct"`
			Source int         `json:"source"`
			ID     string      `json:"id"`
			Name   string      `json:"name"`
		} `json:"values"`
		Name               string      `json:"name"`
		ValueStruct        interface{} `json:"value_struct"`
		AttributeGroupName string      `json:"attribute_group_name"`
		Source             int         `json:"source"`
	} `json:"attributes"`
	DifferentialPricing struct {
		ID int `json:"id"`
	} `json:"differential_pricing,omitempty"`
	OriginalPrice    float64  `json:"original_price"`
	CategoryID       string   `json:"category_id"`
	OfficialStoreID  int      `json:"official_store_id"`
	CatalogProductID string   `json:"catalog_product_id"`
	Tags             []string `json:"tags"`
	CatalogListing   bool     `json:"catalog_listing"`
}

//Category representa GET
type Category struct {
	SiteID string `json:"site_id"`
	Paging struct {
		Total          int `json:"total"`
		Offset         int `json:"offset"`
		Limit          int `json:"limit"`
		PrimaryResults int `json:"primary_results"`
	} `json:"paging"`

	Results          []Result      `json:"results"`
	SecondaryResults []interface{} `json:"secondary_results"`
	RelatedResults   []interface{} `json:"related_results"`
	Sort             struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"sort"`
	AvailableSorts []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"available_sorts"`
	Filters []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Values []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			PathFromRoot []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"path_from_root"`
		} `json:"values"`
	} `json:"filters"`
	AvailableFilters []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Values []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Results int    `json:"results"`
		} `json:"values"`
	} `json:"available_filters"`
}

//Scraping procura quantidade em estoque no anuncio ML
func Scraping(obj ...Result) <-chan string {
	c := make(chan string)
	for _, data := range obj {
		go func(data Result) {
			res, _ := http.Get(data.Permalink)
			doc, _ := goquery.NewDocumentFromReader(res.Body)

			avaliable := doc.Find(".ui-pdp-buybox__quantity__available").First().Text()

			sold := doc.Find(".ui-pdp-subtitle").First().Text()

			rep := strings.NewReplacer("vendidos", "",
				"vendido", "",
				"(", "",
				")", "",
				"Novo  |  ", "",
				"Novo", "0",
				" ", "")
			sold = rep.Replace(sold)

			re := regexp.MustCompile(avaliable)
			if re.Match([]byte(`Último disponível!`)) {
				avaliable = "1"
			} else {
				rep := strings.NewReplacer("disponível", "", "disponíveis", "", "(", "", ")", "", " ", "")
				avaliable = rep.Replace(avaliable)

			}
			av, _ := strconv.Atoi(avaliable)
			s, _ := strconv.Atoi(sold)

			var jsonData []byte

			anuncio := Anuncio{Quantidade: av, Vendido: s, Titulo: data.Title, Permalink: data.Permalink}
			jsonData, _ = json.Marshal(anuncio)

			c <- string(jsonData)

		}(data)
	}
	return c
}

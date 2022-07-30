package main

import (
	"fmt"

	"encoding/json"

	"strconv"

	"strings"

	"github.com/gocolly/colly/v2"
)

type vehicle struct {
	producer                   string
	model                      string
	production_year            int
	traction_type              string
	vehicle_registration_plate string
	vehicle_number             string
	operator                   string
	garage                     string
	ticket_machine             string
	equipment                  string
}

const vehicleStructFieldCount = 10

//here there will be lists of correct sets of values for searchQuery
var traction_types = [4]string{"Autobus", "Kolej", "Metro", "Tramwaj"}

type searchQuery struct {
	traction_type int
	//producer                   string
	//model                      string
	//production_year            int
	//vehicle_registration_plate string
	//vehicle_number             string
	//operator                   string
	//garage                     string
	//in future there will be also other criteria added (these which refers to the equipment)
}

func getPagesNum(url string) int {
	result := 1
	c := colly.NewCollector(
		// Visit only domains:
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	p := 0
	c.OnHTML(".page-numbers>a", func(e *colly.HTMLElement) {
		if p == 1 {
			localResult, error := strconv.Atoi(strings.Split(e.Text, " ")[0])
			fmt.Println(error)
			result = localResult
		}
		p++
	})
	c.Visit(url)
	return result
}

func vehicleStringToInt(input string) int {
	if input == "" {
		return 0
	} else {
		result, error := strconv.Atoi(input)
		fmt.Println(error)
		return result
	}
}

func vehicleToJSON(inputVehicle vehicle) {
	result, error := json.Marshal(inputVehicle)
	fmt.Println(error)
	fmt.Println(result)
}

//it will return array of vehicle numbers found and/or message if needed (with an error for example or sth idk)
func search(searchQuery searchQuery) {
	searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_traction=%d&ztm_make=&ztm_model=&ztm_year=&ztm_registration=&ztm_vehicle_number=&ztm_carrier=&ztm_depot=", searchQuery.traction_type)
	pagesNum := getPagesNum(searchURL)
	resultVehicles := make([]string, 0)
	fmt.Println(pagesNum)
	for i := 0; i < pagesNum; i++ {
		searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/page/%d/?ztm_traction=%d&ztm_make=&ztm_model=&ztm_year=&ztm_registration=&ztm_vehicle_number=&ztm_carrier=&ztm_depot=", i, searchQuery.traction_type)
		fmt.Println(searchURL)
		c := colly.NewCollector(
			// Visit only domains:
			colly.AllowedDomains("www.ztm.waw.pl"),
		)
		p := 0
		c.OnHTML("div[role=cell]", func(e *colly.HTMLElement) {
			if p%5 == 0 {
				vehicleNum := e.Text
				resultVehicles = append(resultVehicles, vehicleNum)
			}
			p++
		})
		c.Visit(searchURL)
	}
	for i := 0; i < len(resultVehicles); i++ {
		fmt.Println(resultVehicles[i])
	}
}
func getVehicleByNum(vehicleNum int) vehicle {
	var retrievedData [10]string

	//get data from website and insert it into array
	vehicleURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_mode=2&ztm_vehicle=%d", vehicleNum)
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains:
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	dataIndex := 0
	c.OnHTML(".vehicle-details-entry-value", func(e *colly.HTMLElement) {
		text := e.Text
		retrievedData[dataIndex] = text
		dataIndex++
	})
	c.Visit(vehicleURL)

	retrievedVehicle := vehicle{
		producer:                   retrievedData[0],
		model:                      retrievedData[1],
		production_year:            vehicleStringToInt(retrievedData[2]),
		traction_type:              retrievedData[3],
		vehicle_registration_plate: retrievedData[4],
		vehicle_number:             retrievedData[5],
		operator:                   retrievedData[6],
		garage:                     retrievedData[7],
		ticket_machine:             retrievedData[8],
		equipment:                  retrievedData[9],
	}

	return retrievedVehicle
}
func main() {
	//fmt.Println("Hello")
	//vehicle := getVehicleByNum(3180)
	//fmt.Println(vehicle)
	examplesearchquery := searchQuery{
		traction_type: 1,
	}
	search(examplesearchquery)
}

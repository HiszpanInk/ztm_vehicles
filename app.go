package main

import (
	"fmt"

	"encoding/json"

	"strconv"

	"strings"

	"github.com/gocolly/colly/v2"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

//this will be used in JSONs
type listElement struct {
	id   int
	name string
}

//this function gets data for the lists with producers, models etc.
//it runs at the start of the program
//var traction_types []string
//var producers []string

func getDataLists() ([]string, []string, []string, []string, []string, []string) {
	traction_types_temp, producers_temp, models_temp, production_years_temp, operators_temp, garages_temp := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)

	url := "https://www.ztm.waw.pl/baza-danych-pojazdow/"
	c := colly.NewCollector(
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	c.OnHTML("#ztm_vehicles_filter_traction > option", func(e *colly.HTMLElement) {
		traction_types_temp = append(traction_types_temp, e.Text)
		fmt.Println(e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_make > option", func(e *colly.HTMLElement) {
		producers_temp = append(producers_temp, e.Text)
		fmt.Println(e.Text)
	})
	c.Visit(url)

	traction_types_temp = append([]string{""}, traction_types_temp[1:]...)
	producers_temp = append([]string{""}, producers_temp[1:]...)
	models_temp = append([]string{""}, models_temp[1:]...)
	production_years_temp = append([]string{""}, production_years_temp[1:]...)
	operators_temp = append([]string{""}, operators_temp[1:]...)
	garages_temp = append([]string{""}, garages_temp[1:]...)

	return traction_types_temp, producers_temp, models_temp, production_years_temp, operators_temp, garages_temp
}

type vehicle struct {
	db_id                      string
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

//the reason searchQuery have diffrent value types than vehicle is that in the URL some values are supposed to be ID's (garages for example)
//those ID's will be based on the list of those elements that are made at the beginning of the program by getDataLists function
type searchQuery struct {
	traction_type              int
	producer                   string
	model                      int
	production_year            int
	vehicle_registration_plate string
	vehicle_number             string
	operator                   int
	garage                     int
	//in future there will be also other criteria added (these which refers to the equipment)
}

//it returns a slice of vehicle numbers found and/or message if needed (with an error for example or sth idk)
func search(searchQuery searchQuery) []string {
	searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_traction=%d&ztm_make=%s&ztm_model=%d&ztm_year=%d&ztm_registration=%s&ztm_vehicle_number=%s&ztm_carrier=%d&ztm_depot=%d", searchQuery.traction_type, searchQuery.producer, searchQuery.model, searchQuery.production_year, searchQuery.vehicle_registration_plate, searchQuery.vehicle_number, searchQuery.operator, searchQuery.garage)
	pagesNum := getPagesNum(searchURL)
	resultVehicles := make([]string, 0)

	for i := 0; i < pagesNum; i++ {
		searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/page/%d/?ztm_traction=%d&ztm_make=%s&ztm_model=%d&ztm_year=%d&ztm_registration=%s&ztm_vehicle_number=%s&ztm_carrier=%d&ztm_depot=%d", i, searchQuery.traction_type, searchQuery.producer, searchQuery.model, searchQuery.production_year, searchQuery.vehicle_registration_plate, searchQuery.vehicle_number, searchQuery.operator, searchQuery.garage)
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
	return resultVehicles
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

func getVehicleByNum(vehicleNum string) vehicle {
	searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_traction=&ztm_make=&ztm_model=&ztm_year=&ztm_registration=&ztm_vehicle_number=%s&ztm_carrier=&ztm_depot=", vehicleNum)
	vehicleURL := ""
	vehicleID := ""
	c2 := colly.NewCollector(
		// Visit only domains:
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	c2.OnHTML(".grid-row-active", func(e *colly.HTMLElement) {
		text := e.Attr("href")
		vehicleURL = text

		vehicleID = reverse(vehicleURL)
		vehicleID = reverse(vehicleID[0:(strings.Index(vehicleID, "="))])
	})
	c2.Visit(searchURL)
	if vehicleURL == "" {
		var emptyVehicle vehicle
		return emptyVehicle
	} else {
		var retrievedData [10]string

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
			db_id:                      vehicleID,
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

}
func main() {
	//this line gets
	//traction_types, producers, models, production_years, operators, garages := getDataLists()

	//fmt.Println("Hello")
	vehicle := getVehicleByNum("2137")

	fmt.Println(vehicle.db_id)
	/*examplesearchquery := searchQuery{
		producer: "Alstom",
	}
	search(examplesearchquery)*/

}

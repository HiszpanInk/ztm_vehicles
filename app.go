package main

import (
	"fmt"

	"strconv"

	"github.com/gocolly/colly/v2"
)

type vehicle struct {
	producer                   string
	model                      string
	production_year            int
	traction_type              string
	vehicle_registration_plate string
	vehicle_number             int
	operator                   string
	garage                     string
	ticket_machine             string
	equipment                  string
}

const vehicleStructFieldCount = 10

func vehicleStringToInt(input string) int {
	if input == "" {
		return 0
	} else {
		result, error := strconv.Atoi(input)
		fmt.Println(error)
		return result
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
		vehicle_number:             vehicleStringToInt(retrievedData[5]),
		operator:                   retrievedData[6],
		garage:                     retrievedData[7],
		ticket_machine:             retrievedData[8],
		equipment:                  retrievedData[9],
	}

	return retrievedVehicle
}
func main() {
	fmt.Println("Hello")
	vehicle := getVehicleByNum(3180)
	fmt.Println(vehicle)
}

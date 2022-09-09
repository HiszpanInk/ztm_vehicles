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

func getElementIndexInSlice(element string, slice []string) int {
	toReturn := -1
	for i := 0; i < len(slice); i++ {
		if element == slice[i] {
			toReturn = i
		}
	}
	return toReturn
}

// this is used for Producers, garages, operators and models
type propertyWithID struct {
	Id   int
	Name string
}

// this function gets data for the lists with Producers, models etc.
// it runs at the start of the program
// var traction_types []string
var producers []string
var models []string
var traction_types []string
var operators []string
var production_years []string
var garages []string

func getDataLists() ([]string, []string, []string, []string, []string, []string) {
	traction_types_temp, producers_temp, models_temp, production_years_temp, operators_temp, garages_temp := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)

	url := "https://www.ztm.waw.pl/baza-danych-pojazdow/"
	c := colly.NewCollector(
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	c.OnHTML("#ztm_vehicles_filter_traction > option", func(e *colly.HTMLElement) {
		traction_types_temp = append(traction_types_temp, e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_make > option", func(e *colly.HTMLElement) {
		producers_temp = append(producers_temp, e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_model > option", func(e *colly.HTMLElement) {
		models_temp = append(models_temp, e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_year > option", func(e *colly.HTMLElement) {
		production_years_temp = append(production_years_temp, e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_carrier > option", func(e *colly.HTMLElement) {
		operators_temp = append(operators_temp, e.Text)
	})
	c.OnHTML("#ztm_vehicles_filter_depot > option", func(e *colly.HTMLElement) {
		garages_temp = append(garages_temp, e.Text)
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

func returnDataLists() DataListsQueryResult {
	if len(producers) == 0 || len(models) == 0 || len(traction_types) == 0 || len(operators) == 0 || len(production_years) == 0 || len(garages) == 0 {
		result := DataListsQueryResult{
			SearchResult: SearchResult{
				Message:       "lists are partly or fully unavailable, please contact administrator",
				Results_count: 0,
			},
		}
		return result
	} else {
		result := DataListsQueryResult{
			SearchResult: SearchResult{
				Message:       "ok",
				Results_count: (len(producers) + len(models) + len(traction_types) + len(operators) + len(production_years) + len(garages)),
			},
			Data: make(map[string][]propertyWithID),
		}

		result.Data["producers"] = stringListToPropertyWithIDList(producers)
		result.Data["models"] = stringListToPropertyWithIDList(models)
		result.Data["traction_types"] = stringListToPropertyWithIDList(traction_types)
		result.Data["operators"] = stringListToPropertyWithIDList(operators)
		result.Data["production_years"] = stringListToPropertyWithIDList(production_years)
		result.Data["garages"] = stringListToPropertyWithIDList(garages)
		return result
	}
}

func stringListToPropertyWithIDList(inputList []string) []propertyWithID {
	var outputList []propertyWithID
	for i := 0; i < len(inputList); i++ {
		outputList = append(outputList, propertyWithID{(i + 1), inputList[i]})
	}
	return outputList
}

type vehicle struct {
	Db_id                      string
	Producer                   propertyWithID
	Model                      propertyWithID
	Production_year            int
	Traction_type              propertyWithID
	Vehicle_registration_plate string
	Vehicle_number             string
	Operator                   propertyWithID
	Garage                     propertyWithID
	Ticket_machine             string
	Equipment                  string
}

//const vehicleStructFieldCount = 10

// the reason searchQuery have diffrent value types than vehicle is that in the URL some values are supposed to be ID's (garages for example)
// those ID's will be based on the list of those elements that are made at the beginning of the program by getDataLists function
type searchQuery struct {
	traction_type              int
	producer                   int
	model                      int
	production_year            int
	vehicle_registration_plate string
	vehicle_number             string
	operator                   int
	garage                     int
	//in future there will be also other criteria added (these which refers to the equipment)
}
type SearchResult struct {
	Message       string
	Results_count int
}
type VehicleSearchQueryResult struct {
	SearchResult
	Data []vehicle
}
type DataListsQueryResult struct {
	SearchResult
	Data map[string][]propertyWithID
}

// it returns a slice of vehicle numbers found and/or message if needed (with an error for example or sth idk)
func search(searchQuery searchQuery, onlyID bool) VehicleSearchQueryResult {

	searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_traction=%d&ztm_make=%d&ztm_model=%d&ztm_year=%d&ztm_registration=%s&ztm_vehicle_number=%s&ztm_carrier=%d&ztm_depot=%d", searchQuery.traction_type, searchQuery.producer, searchQuery.model, searchQuery.production_year, searchQuery.vehicle_registration_plate, searchQuery.vehicle_number, searchQuery.operator, searchQuery.garage)
	pagesNum := getPagesNum(searchURL)
	//resultVehicles := make([]string, 0)

	var vehicleURL []string
	var vehicleID []string

	var vehicleURLTemp string
	var vehicleIDTemp string

	for i := 1; i < pagesNum; i++ {
		searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/page/%d/?ztm_traction=%d&ztm_make=%d&ztm_model=%d&ztm_year=%d&ztm_registration=%s&ztm_vehicle_number=%s&ztm_carrier=%d&ztm_depot=%d", i, searchQuery.traction_type, searchQuery.producer, searchQuery.model, searchQuery.production_year, searchQuery.vehicle_registration_plate, searchQuery.vehicle_number, searchQuery.operator, searchQuery.garage)
		fmt.Println(searchURL)
		c := colly.NewCollector(
			// Visit only domains:
			colly.AllowedDomains("www.ztm.waw.pl"),
		)
		c.OnHTML(".grid-row-active", func(e *colly.HTMLElement) {
			text := e.Attr("href")
			vehicleURLTemp = text

			vehicleURL = append(vehicleURL, vehicleURLTemp)

			vehicleIDTemp = reverse(vehicleURLTemp)
			vehicleIDTemp = reverse(vehicleIDTemp[0:(strings.Index(vehicleIDTemp, "="))])

			vehicleID = append(vehicleID, vehicleIDTemp)
		})
		c.Visit(searchURL)
	}
	var resultVehicles []vehicle
	if onlyID {
		for _, element := range vehicleID {
			tempVehicle := vehicle{
				Db_id: element,
			}
			resultVehicles = append(resultVehicles, tempVehicle)
		}
	} else {
		for _, element := range vehicleID {
			tempVehicle := getVehicleData(element)
			resultVehicles = append(resultVehicles, tempVehicle)
		}
	}
	result := VehicleSearchQueryResult{
		SearchResult: SearchResult{
			Message:       "ok",
			Results_count: len(vehicleID),
		},
		Data: resultVehicles,
	}
	return result
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
			thirdPageNumberLink := strings.Split(e.Text, " ")[0]
			if thirdPageNumberLink != "Następna" {
				localResult, _ := strconv.Atoi(strings.Split(e.Text, " ")[0])
				//fmt.Println(error)
				result = localResult
			} else {
				result = 2
			}
		}
		p++
	})
	c.Visit(url)
	result++
	return result
}

func vehicleStringToInt(input string) int {
	if input == "" {
		return 0
	} else {
		result, _ := strconv.Atoi(input)
		//fmt.Println(error)
		return result
	}
}

func vehicleToJSON(input VehicleSearchQueryResult) []byte {
	result, _ := json.MarshalIndent(&input, "", "  ")
	//fmt.Println(error)
	return result
}

func getVehicleData(vehicleID string) vehicle {
	var retrievedData [10]string
	vehicleURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_mode=2&ztm_vehicle=%s", vehicleID)
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

	if dataIndex == 0 {
		var emptyVehicle vehicle
		return emptyVehicle
	} else {
		resultVehicle := vehicle{
			Db_id:                      vehicleID,
			Producer:                   propertyWithID{getElementIndexInSlice(retrievedData[0], producers), retrievedData[0]},
			Model:                      propertyWithID{getElementIndexInSlice(retrievedData[1], models), retrievedData[1]},
			Production_year:            vehicleStringToInt(retrievedData[2]),
			Traction_type:              propertyWithID{getElementIndexInSlice(retrievedData[3], traction_types), retrievedData[3]},
			Vehicle_registration_plate: retrievedData[4],
			Vehicle_number:             retrievedData[5],
			Operator:                   propertyWithID{getElementIndexInSlice(retrievedData[6], operators), retrievedData[6]},
			Garage:                     propertyWithID{getElementIndexInSlice(retrievedData[7], garages), retrievedData[7]},
			Ticket_machine:             retrievedData[8],
			Equipment:                  retrievedData[9],
		}
		return resultVehicle
	}
}

func getVehicleById(vehicleID string) VehicleSearchQueryResult {
	tempVehicle := getVehicleData(vehicleID)

	if tempVehicle.Db_id == "0" {
		var emptyVehicles []vehicle
		result := VehicleSearchQueryResult{
			SearchResult: SearchResult{
				Message:       "no vehicle found",
				Results_count: 0,
			},
			Data: emptyVehicles,
		}
		return result
	} else {
		var resultVehicle []vehicle
		resultVehicle = append(resultVehicle, tempVehicle)
		result := VehicleSearchQueryResult{
			SearchResult: SearchResult{
				Message:       "ok",
				Results_count: 1,
			},
			Data: resultVehicle,
		}
		return result
	}
}

func getVehicleByNum(vehicleNum string) VehicleSearchQueryResult {
	searchURL := fmt.Sprintf("https://www.ztm.waw.pl/baza-danych-pojazdow/?ztm_traction=&ztm_make=&ztm_model=&ztm_year=&ztm_registration=&ztm_vehicle_number=%s&ztm_carrier=&ztm_depot=", vehicleNum)

	var vehicleURL []string
	var vehicleID []string

	var vehicleURLTemp string
	var vehicleIDTemp string

	c2 := colly.NewCollector(
		// Visit only domains:
		colly.AllowedDomains("www.ztm.waw.pl"),
	)
	c2.OnHTML(".grid-row-active", func(e *colly.HTMLElement) {
		text := e.Attr("href")
		vehicleURLTemp = text

		vehicleURL = append(vehicleURL, vehicleURLTemp)

		vehicleIDTemp = reverse(vehicleURLTemp)
		vehicleIDTemp = reverse(vehicleIDTemp[0:(strings.Index(vehicleIDTemp, "="))])

		vehicleID = append(vehicleID, vehicleIDTemp)
	})
	c2.Visit(searchURL)
	if len(vehicleURL) == 0 {
		var emptyVehicles []vehicle
		result := VehicleSearchQueryResult{
			SearchResult: SearchResult{
				Message:       "no vehicle found",
				Results_count: 0,
			},
			Data: emptyVehicles,
		}
		return result
	} else {
		var retrievedVehicles []vehicle
		for _, element := range vehicleID {
			tempVehicle := getVehicleData(element)
			retrievedVehicles = append(retrievedVehicles, tempVehicle)
		}
		result := VehicleSearchQueryResult{
			SearchResult: SearchResult{
				Message:       "ok",
				Results_count: len(retrievedVehicles),
			},
			Data: retrievedVehicles,
		}

		return result
	}

}
func main() {
	//for now main() part is used only for testing

	traction_types, producers, models, production_years, operators, garages = getDataLists()
	result, _ := json.MarshalIndent((returnDataLists()), "", "  ")
	fmt.Println(string(result))
	/*
			//fmt.Println("Hello")
			vehicle := getVehicleByNum("2022")
			fmt.Println(string(vehicleToJSON(vehicle)))
			vehicle = getVehicleById("2137")
			fmt.Println(string(vehicleToJSON(vehicle)))

		examplesearchquery := searchQuery{
			production_year: 2022,
		}
		search(examplesearchquery, true)

		examplesearchquery2 := searchQuery{
			operator: 2,
		}
		search(examplesearchquery2, true)*/
	//search(examplesearchquery, true)

}

package groupietrackers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var FakeCurrentYear int
var FakeCurrentMonth time.Month
var FakeCurrentDay int

func GetAPIData(apiUrl string) []byte { // ! Do a request to an API and return it content
	apiClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("User", "groupie-tracker")
	res, getErr := apiClient.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(readErr)
	}

	return body

}

func SetGlobalData(body []byte) ApiData {
	apidata := ApiData{}
	jsonErr := json.Unmarshal(body, &apidata)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return apidata
}

func GetGeoCodeData(body []byte) map[string]interface{} {  // ! Convert Google Map API data for golang
	ap := make(map[string]interface{})
	jsonErr := json.Unmarshal(body, &ap)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return ap
}

func GetCoord(geoCodeData map[string]interface{}) GeoCoord {  // ! Get longitude and latitude in Google Map API data
	coord := GeoCoord{}
	x := geoCodeData["results"].([]interface{})
	y := x[0].(map[string]interface{})
	z := y["geometry"].(map[string]interface{})
	y = z["location"].(map[string]interface{})
	coord.Lat = y["lat"].(float64)
	coord.Long = y["lng"].(float64)
	return coord
}

func SetArtistData(body []byte) artistStruct {  // ! Convert Artist API data for golang
	artistdata := artistStruct{}
	jsonErr := json.Unmarshal(body, &artistdata)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return artistdata
}

func SetLocationData(body []byte) locationsStruct {  // ! Convert Location API data for golang
	locationdata := locationsStruct{}
	jsonErr := json.Unmarshal(body, &locationdata)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return locationdata
}

func SetDateData(body []byte) datesStruct {  // ! Convert Date API data for golang
	datedata := datesStruct{}
	jsonErr := json.Unmarshal(body, &datedata)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return datedata
}

func SetRelationData(body []byte) relationStruct {  // ! Convert Relation API data for golang
	relationdata := relationStruct{}
	jsonErr := json.Unmarshal(body, &relationdata)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return relationdata
}

func SetArtist(body []byte) []Artist {  // ! Convert Artist API data for golang
	artistd := []Artist{}
	jsonErr := json.Unmarshal(body, &artistd)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return artistd
}

func UpdateCurrentBand(apiArtist string) CurrentBand {  // ! Get data of band select by user and format them
	dataartist := SetArtistData(GetAPIData(apiArtist))
	datadatelocation := SetRelationData(GetAPIData(dataartist.Relations))

	cb := CurrentBand{
		Id:           dataartist.Id,
		Name:         dataartist.Name,
		Image:        dataartist.Image,
		Member:       dataartist.Member,
		CreationDate: dataartist.CreationDate,
		FirstAlbum:   dataartist.FirstAlbum,
		Relations:    ChangeDateFormat(datadatelocation.DatesLocations),
	}
	cb.FuturRelation, cb.PassRelation = CheckRelationTime(cb.Relations)
	return cb
}

func ChangeDateFormat(date map[string][]string) map[string][][][]string { // ! Format date for a readable format
	nDate := make(map[string][][][]string)

	for location, ldate := range date {
		var ville string
		var pays string

		if len(ldate) > 1 {
			lDate := [][]string{}
			var allDate [][]string
			for _, d := range ldate {
				allDate = append(allDate, strings.Split(d, "-"))
			}
			var day string
			var mday string
			var lday int
			var month string
			var year string

			pays = strings.Split(location, "-")[1]
			ville = strings.Split(location, "-")[0]

			for i, d := range allDate {
				if i == 0 {
					day = d[0]
					mday = d[0]
					lday, _ = strconv.Atoi(day)
					month = d[1]
					year = d[2]
				} else {
					nlday, _ := strconv.Atoi(d[0])
					if month == d[1] && year == d[2] && lday-1 == nlday {
						day = d[0] + "-" + mday
						lday--
						if i == len(allDate)-1 {
							imonth, _ := strconv.Atoi(month)
							lDate = append(lDate, []string{ville, day, []string{"Jan", "Feb", "Mar", "Apr", "May", "June", "July", "Aug", "Sept", "Oct", "Nov", "Dec"}[imonth-1], year})
						}
					} else {
						imonth, _ := strconv.Atoi(month)
						lDate = append(lDate, []string{ville, day, []string{"Jan", "Feb", "Mar", "Apr", "May", "June", "July", "Aug", "Sept", "Oct", "Nov", "Dec"}[imonth-1], year})
						day = d[0]
						mday = d[0]
						lday, _ = strconv.Atoi(day)
						month = d[1]
						year = d[2]
					}
				}
			}

			nDate[pays] = append(nDate[pays], lDate)
		} else {
			pays = strings.Split(location, "-")[1]
			ville = strings.Split(location, "-")[0]

			lDate := [][]string{}
			imonth, _ := strconv.Atoi(strings.Split(ldate[0], "-")[1])
			lDate = append(lDate, []string{ville, strings.Split(ldate[0], "-")[0], []string{"Jan", "Feb", "Mar", "Apr", "May", "June", "July", "Aug", "Sept", "Oct", "Nov", "Dec"}[imonth-1], strings.Split(ldate[0], "-")[2]})
			nDate[pays] = append(nDate[pays], lDate)
		}
	}
	return nDate
}

func SetCoordToEvent(event map[string][]Event) map[string][]Event {  // ! Get long and lat of a location to be use in Google Map API
	for pays := range event {
		for i, e := range event[pays] {
			e.Coord = GetCoord(GetGeoCodeData(GetAPIData("https://maps.googleapis.com/maps/api/geocode/json?address=" + e.City + "+" + e.Country + "&key=AIzaSyBq9H9P3Jazc6tUoqQ8fwBdMbgLhm0QSe4")))
			event[pays][i] = FormatFLocation(e)
		}
	}
	return event
}

func FormatFLocation(event Event) Event {  // ! Format location to be readable
	city := strings.Split(event.City, "_")
	for i, value := range city {
		ncity := ""
		for j, lettre := range value {
			if j == 0 {
				ncity += string(lettre - 32)
			} else {
				ncity += string(lettre)
			}
		}
		city[i] = ncity
	}
	event.City = strings.Join(city, " ")

	country := strings.Split(event.Country, "_")
	for i, value := range country {
		ncountry := ""
		for j, lettre := range value {
			if j == 0 {
				ncountry += string(lettre - 32)
			} else {
				ncountry += string(lettre)
			}
		}
		country[i] = ncountry
	}
	event.Country = strings.Join(country, " ")

	return event
}

func CheckRelationTime(date map[string][][][]string) (map[string][]Event, [][]string) {  // ! Check if the relation is passed

	FakeCurrentYear, FakeCurrentMonth, FakeCurrentDay = time.Now().Date()
	FakeCurrentYear -= 3
	fRelation := make(map[string][]Event)
	pRelation := [][]string{}

	for pays := range date {
		for _, location := range date[pays] {
			for _, rlocation := range location {
				checkEvent := Event{}
				switch {
				case AtoiWithoutErr(rlocation[3]) >= FakeCurrentYear:
					checkEvent.Country = pays
					checkEvent.City = rlocation[0]
					checkEvent.Date = rlocation[1:]
					fRelation[pays] = append(fRelation[pays], checkEvent)
				case AtoiWithoutErr(rlocation[3]) < FakeCurrentYear:
					rlocation = append(rlocation, pays)
					pRelation = append(pRelation, rlocation)
				default:
				}
			}
		}
	}

	pRelation = sortByNIndex(3, sortByNIndex(1, pRelation))
	if len(pRelation) >= 3 {
		pRelation = pRelation[:3]
	}
	/*fRelation = FormatFLocation(fRelation)*/
	fRelation = SetCoordToEvent(fRelation)
	return fRelation, pRelation
}

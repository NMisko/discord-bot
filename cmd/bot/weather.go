/*	Methods that get weather information.
	TODO: Hide the API Key
*/

package main

import (
    "fmt"
    "net/http"
    "io/ioutil"

    log "github.com/Sirupsen/logrus"
    "encoding/xml"
)

var (
    KEY string = "babe442a111d8e2dc0d9b9145a6fc1ae"
)

//layout: http://api.openweathermap.org/data/2.5/forecast/daily?q=stuttgart&mode=xml&units=metric&cnt=7&appid=babe442a111d8e2dc0d9b9145a6fc1ae

/*	Struct containing the structure to unmarshal the data returned by the Openweather API.
*/
type Weather struct {
    Forecast struct {
        Time []struct {
            Symbol struct {
                Name string `xml:"name,attr"`
            } `xml:"symbol"`
            Temperature struct {
                Unit string `xml:"unit,attr"`
                Value string `xml:"value,attr"`
                Min string `xml:"min,attr"`
                Max string `xml:"max,attr"`
            } `xml:"temperature"`
            WindDirection struct {
                 Name string `xml:"name,attr"`
            } `xml:"windDirection"`
            WindSpeed struct {
                Name string `xml:"name,attr"`
            } `xml:"windSpeed"`
        } `xml:"time"`
    } `xml:"forecast"`
}

/*	Returns a 'Weather' struct containing information about the weather in the given city/country.
*/
func GetWeather(city string, country string) Weather {
    resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast/daily?q=%s,%s&mode=xml&units=metric&cnt=7&appid=%s", city, country, KEY))
    if err != nil {
            log.WithFields(log.Fields{
                "error": err,
            }).Warning("Failed to GET city: ")
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    var w Weather
    xml.Unmarshal(body, &w)

    log.Info(w)

    return w
}

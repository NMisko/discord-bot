package main

import (
    "fmt"
    "net/http"
    "io/ioutil"

    log "github.com/Sirupsen/logrus"
    "encoding/xml"
)
//layout: http://api.openweathermap.org/data/2.5/forecast/daily?q=stuttgart&mode=xml&units=metric&cnt=7&appid=babe442a111d8e2dc0d9b9145a6fc1ae
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

func GetWeather(city string, country string) Weather {
    KEY := "babe442a111d8e2dc0d9b9145a6fc1ae"
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
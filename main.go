package main

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

const API = "https://api.sunrise-sunset.org/json?lat=52.09755&lng=23.68775&formatted=0"

type SunResults struct {
	Sun
	error
}

var instance *SunResults

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", health)
	e.GET("/", status)
	e.GET("/sun", sun)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, BaseResponse{
		Message: http.StatusText(http.StatusOK),
	})
}

func sun(c echo.Context) error {
	var sun = *getSun()
	if sun.error != nil {
		return c.JSON(http.StatusInternalServerError, BaseResponse{
			Message: sun.error.Error(),
		})
	} else {
		return c.JSON(http.StatusOK, sun.Sun)
	}
}

func status(c echo.Context) error {
	var s = *getSun()
	if s.error != nil {
		return c.JSON(http.StatusInternalServerError, BaseResponse{
			Message: s.error.Error(),
		})
	} else {
		var curTime = getUTC().Unix()
		return c.JSON(http.StatusOK, Status{
			Twilight:             curTime <= s.Results.Sunrise.Unix() || curTime >= s.Results.Sunset.Unix(),
			CivilTwilight:        curTime <= s.Results.CivilTwilightBegin.Unix() || curTime >= s.Results.CivilTwilightEnd.Unix(),
			NauticalTwilight:     curTime <= s.Results.NauticalTwilightBegin.Unix() || curTime >= s.Results.NauticalTwilightEnd.Unix(),
			AstronomicalTwilight: curTime <= s.Results.AstronomicalTwilightBegin.Unix() || curTime >= s.Results.AstronomicalTwilightEnd.Unix(),
		})
	}
}

type Sun struct {
	Results struct {
		Sunrise                   time.Time `json:"sunrise"`
		Sunset                    time.Time `json:"sunset"`
		SolarNoon                 time.Time `json:"solar_noon"`
		DayLength                 int       `json:"day_length"`
		CivilTwilightBegin        time.Time `json:"civil_twilight_begin"`
		CivilTwilightEnd          time.Time `json:"civil_twilight_end"`
		NauticalTwilightBegin     time.Time `json:"nautical_twilight_begin"`
		NauticalTwilightEnd       time.Time `json:"nautical_twilight_end"`
		AstronomicalTwilightBegin time.Time `json:"astronomical_twilight_begin"`
		AstronomicalTwilightEnd   time.Time `json:"astronomical_twilight_end"`
	} `json:"results"`
	Status string `json:"status"`
}

type Status struct {
	Twilight             bool `json:"twilight"`
	CivilTwilight        bool `json:"civil_twilight"`
	NauticalTwilight     bool `json:"nautical_twilight"`
	AstronomicalTwilight bool `json:"astronomical_twilight"`
}

type BaseResponse struct {
	Message string `json:"message"`
}

func getSun() *SunResults {
	if instance == nil || instance.error != nil || getUTC().Day() > instance.Sun.Results.Sunrise.Day() {
		setSun()
	}
	return instance
}

func setSun() {
	sun, err := parseSun()
	instance = &SunResults{Sun: sun, error: err}
}

func parseSun() (sun Sun, err error) {
	err = getJSON(API, &sun)
	return
}

func getUTC() time.Time {
	return time.Now().UTC()
}

func getJSON(url string, result interface{}) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return
	}
	return nil
}

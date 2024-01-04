package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yomorun/yomo"
	"github.com/yomorun/yomo/serverless"

	"demo"
)

type Msg struct {
	CityName string `json:"city_name"`
}

type WeatherResponse struct {
	Current struct {
		TempC      float64 `json:"temp_c"`
		FeelslikeC float64 `json:"feelslike_c"`
	} `json:"current"`
}

func getTemperature(city string) (float64, float64, error) {
	apiKey := os.Getenv("API_KEY")
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var weatherResponse WeatherResponse
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		return 0, 0, err
	}

	return weatherResponse.Current.TempC, weatherResponse.Current.FeelslikeC, nil
}

func Handler(ctx serverless.Context) {
	var msg Msg
	err := json.Unmarshal(ctx.Data(), &msg)
	if err != nil {
		println("error: ", msg.CityName)
		ctx.Write(0x30, []byte("error: json unmarshal error: "+err.Error()))
		return
	}

	println("get-weather: ", msg.CityName)

	city := msg.CityName
	temp, feelslike, err := getTemperature(city)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	println("get-weather from API: ", temp, feelslike)

	res := fmt.Sprintf("[cc|weatherapi.com] ok: the temperature of %s is %.1f℃, feel like %.1f℃", msg.CityName, temp, feelslike)
	ctx.Write(0x30, []byte(res))
}

func main() {
	sfn := yomo.NewStreamFunction("get-weather", demo.ZipperAddr)
	sfn.SetObserveDataTags(0x32)
	sfn.SetHandler(Handler)
	err := sfn.Connect()
	if err != nil {
		log.Fatalln(err)
	}
	defer sfn.Close()
	sfn.Wait()
}

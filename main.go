package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	defaultAPIKey = "" // Set via WEATHER_API_KEY env var
	baseURL       = "https://api.openweathermap.org/data/2.5"
)

type WeatherData struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
}

type ForecastData struct {
	List []struct {
		DtTxt string `json:"dt_txt"`
		Main  struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Pop float64 `json:"pop"` // probability of precipitation
	} `json:"list"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

func getAPIKey() string {
	key := os.Getenv("WEATHER_API_KEY")
	if key == "" {
		return defaultAPIKey
	}
	return key
}

func fetchWeather(city string) (string, error) {
	apiKey := getAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("WEATHER_API_KEY not set")
	}

	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=imperial", baseURL, city, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var data WeatherData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse weather: %w", err)
	}

	condition := "Clear"
	description := "clear sky"
	if len(data.Weather) > 0 {
		condition = data.Weather[0].Main
		description = data.Weather[0].Description
	}

	result := fmt.Sprintf(`Weather for %s:
  Condition: %s (%s)
  Temperature: %.0f°F (feels like %.0f°F)
  High: %.0f°F / Low: %.0f°F
  Humidity: %d%%
  Wind: %.1f mph (gusts: %.1f mph)
  Cloud cover: %d%%`,
		data.Name, condition, description,
		data.Main.Temp, data.Main.FeelsLike,
		data.Main.TempMax, data.Main.TempMin,
		data.Main.Humidity,
		data.Wind.Speed, data.Wind.Gust,
		data.Clouds.All)

	return result, nil
}

func fetchForecast(city string, hours int) (string, error) {
	apiKey := getAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("WEATHER_API_KEY not set")
	}

	url := fmt.Sprintf("%s/forecast?q=%s&appid=%s&units=imperial", baseURL, city, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch forecast: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var data ForecastData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse forecast: %w", err)
	}

	// Each entry is 3 hours, so hours/3 entries
	entries := hours / 3
	if entries > len(data.List) {
		entries = len(data.List)
	}
	if entries < 1 {
		entries = 1
	}

	result := fmt.Sprintf("Forecast for %s (next %d hours):\n", data.City.Name, hours)
	for i := 0; i < entries; i++ {
		entry := data.List[i]
		condition := "Clear"
		if len(entry.Weather) > 0 {
			condition = entry.Weather[0].Description
		}
		result += fmt.Sprintf("  %s: %.0f°F, %s, wind %.1f mph, rain chance %.0f%%\n",
			entry.DtTxt, entry.Main.Temp, condition, entry.Wind.Speed, entry.Pop*100)
	}

	return result, nil
}

func main() {
	s := server.NewMCPServer(
		"Weather MCP",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool: get_weather
	s.AddTool(
		mcp.NewTool("get_weather",
			mcp.WithDescription("Get current weather conditions for a city"),
			mcp.WithString("city",
				mcp.Required(),
				mcp.Description("City name (e.g., 'Austin, TX', 'London', 'Tokyo')")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			city, ok := req.Params.Arguments["city"].(string)
			if !ok || city == "" {
				return mcp.NewToolResultError("city is required"), nil
			}

			result, err := fetchWeather(city)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get weather: %v", err)), nil
			}

			return mcp.NewToolResultText(result), nil
		},
	)

	// Tool: get_forecast
	s.AddTool(
		mcp.NewTool("get_forecast",
			mcp.WithDescription("Get weather forecast for upcoming hours"),
			mcp.WithString("city",
				mcp.Required(),
				mcp.Description("City name (e.g., 'Austin, TX', 'London', 'Tokyo')")),
			mcp.WithNumber("hours",
				mcp.Description("Hours to forecast (default 24, max 120)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			city, ok := req.Params.Arguments["city"].(string)
			if !ok || city == "" {
				return mcp.NewToolResultError("city is required"), nil
			}

			hours := 24
			if h, ok := req.Params.Arguments["hours"].(float64); ok && h > 0 {
				hours = int(h)
				if hours > 120 {
					hours = 120
				}
			}

			result, err := fetchForecast(city, hours)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get forecast: %v", err)), nil
			}

			return mcp.NewToolResultText(result), nil
		},
	)

	log.Println("Weather MCP server starting on stdio...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

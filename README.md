# Weather MCP Server

A lightweight MCP server that provides weather data to AI agents. Designed to pair with [LightningProx MCP](https://github.com/unixlamadev-spec/lightningprox-mcp) to demonstrate autonomous agent workflows — where one MCP provides data and another handles payment.

## What This Demonstrates

```
You: "Check the weather in Austin and use LightningProx to analyze if I need a jacket"

Agent autonomously:
  1. Calls Weather MCP → get_weather("Austin, TX")
  2. Calls LightningProx MCP → check_balance()
  3. Calls LightningProx MCP → ask_ai() with weather data
  4. Returns: "It's 45°F and windy — yes, bring a jacket"
```

Two MCP servers. One payment. Zero human intervention.

## Install

```bash
go install github.com/unixlamadev-spec/weather-mcp/cmd/mcp-server@latest
```

## Setup

1. Get a free API key from [OpenWeatherMap](https://openweathermap.org/api) (free tier = 1000 calls/day)

2. Add both MCP servers to Claude Desktop config:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "lightningprox": {
      "command": "mcp-server",
      "args": [],
      "env": {}
    },
    "weather": {
      "command": "weather-mcp-server",
      "args": [],
      "env": {
        "WEATHER_API_KEY": "your_openweathermap_api_key"
      }
    }
  }
}
```

3. Restart Claude Desktop

## Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_weather` | Current conditions | `city` (required) |
| `get_forecast` | Upcoming forecast | `city` (required), `hours` (optional, default 24) |

## Example Prompts

**Basic weather:**
> "What's the weather in Tokyo?"

**Chained with LightningProx:**
> "Get the weather in Austin, then use LightningProx to ask Claude if I should bring an umbrella based on that data."

**Multi-city comparison:**
> "Compare the weather in New York and Miami. Use LightningProx to get an AI recommendation for which city to visit this weekend."

## The Pattern

This isn't just a weather app. It's a reference implementation of **MCP service chaining**:

- **Service MCP** provides data (weather, stocks, news, anything)
- **LightningProx MCP** handles payment for AI analysis
- **Agent** orchestrates both without human involvement

Any API can become an MCP server. LightningProx can pay for any of them.

## License

MIT

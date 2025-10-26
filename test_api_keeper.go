package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gp/projects/alex/dollar/nuahchain_osmosis/x/oracle/sources"
)

func main() {
	fmt.Println("🔍 Тест APIKeeper для оракула")
	fmt.Println("=============================")

	// Создаем APIKeeper (в реальном приложении нужны правильные параметры)
	// apiKeeper := keeper.NewAPIKeeper(cdc, storeKey, authority)

	// Для тестирования создаем API клиент напрямую
	client := sources.NewAPIClient()

	// Тестируем различные символы
	testSymbols := []string{
		"GC=F",     // Золото
		"SI=F",     // Серебро
		"CL=F",     // Нефть
		"EURUSD=X", // EUR/USD
		"AAPL",     // Apple
		"BTC-USD",  // Bitcoin
	}

	fmt.Println("📊 Тестирование получения цен:")
	fmt.Println("=============================")

	for _, symbol := range testSymbols {
		fmt.Printf("\n📈 Символ: %s\n", symbol)
		fmt.Println(strings.Repeat("-", 30))

		// Получаем текущую цену
		priceData, err := client.GetPriceFromYahooFinance(symbol)
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
			continue
		}

		fmt.Printf("✅ Цена: $%.2f\n", priceData.Price)
		fmt.Printf("📊 Изменение: $%.2f (%.2f%%)\n", priceData.Change, priceData.ChangePercent)
		fmt.Printf("🏛️ Биржа: %s\n", priceData.Exchange)
		fmt.Printf("📅 Время: %s\n", priceData.Timestamp.Format("2006-01-02 15:04:05"))

		// Небольшая задержка между запросами
		time.Sleep(client.GetRateLimit())
	}

	// Тестируем исторические данные
	fmt.Println("\n📈 Тестирование исторических данных:")
	fmt.Println("====================================")

	symbol := "GC=F"
	fmt.Printf("📊 Символ: %s\n", symbol)

	historicalData, err := client.GetHistoricalData(symbol, "5d")
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
	} else {
		fmt.Printf("✅ Получено %d исторических точек\n", len(historicalData))

		// Показываем последние 3 точки
		start := len(historicalData) - 3
		if start < 0 {
			start = 0
		}

		for i := start; i < len(historicalData); i++ {
			data := historicalData[i]
			fmt.Printf("  %s: $%.2f (%.2f%%)\n",
				data.Timestamp.Format("2006-01-02"),
				data.Price,
				data.ChangePercent)
		}
	}

	// Тестируем поиск символов
	fmt.Println("\n🔍 Тестирование поиска символов:")
	fmt.Println("===============================")

	searchResults, err := client.SearchSymbols("gold")
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
	} else {
		fmt.Printf("✅ Найдено %d результатов\n", len(searchResults))

		// Показываем первые 3 результата
		for i, result := range searchResults {
			if i >= 3 {
				break
			}
			if symbol, ok := result["symbol"].(string); ok {
				if longName, ok := result["longName"].(string); ok {
					fmt.Printf("  %d. %s - %s\n", i+1, symbol, longName)
				} else {
					fmt.Printf("  %d. %s\n", i+1, symbol)
				}
			}
		}
	}

	// Тестируем категории символов
	fmt.Println("\n📋 Тестирование категорий символов:")
	fmt.Println("===================================")

	categories := sources.GetDefaultSymbols()
	for category, symbols := range categories {
		fmt.Printf("\n📊 %s (%d символов):\n", category, len(symbols))
		for i, symbol := range symbols {
			if i >= 2 { // Показываем только первые 2 символа из каждой категории
				fmt.Printf("  ... и еще %d символов\n", len(symbols)-2)
				break
			}
			priceData, err := client.GetPriceFromYahooFinance(symbol)
			if err != nil {
				fmt.Printf("  ❌ %s: Ошибка\n", symbol)
			} else {
				fmt.Printf("  ✅ %s: $%.2f\n", symbol, priceData.Price)
			}
			time.Sleep(client.GetRateLimit())
		}
	}

	fmt.Println("\n🏁 Тест завершен!")
}


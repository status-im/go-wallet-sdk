package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

func main() {
	fmt.Println("🌐 Token Fetcher Example")
	fmt.Println("=========================")

	// Create fetcher instance
	tokenFetcher := fetcher.New(fetcher.DefaultConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example 1: Fetch single token list
	fmt.Println("\n📋 Single Token List Fetch")
	fmt.Println("============================")
	demonstrateSingleFetch(ctx, tokenFetcher)

	// Example 2: Fetch multiple token lists concurrently
	fmt.Println("\n🚀 Concurrent Token List Fetch")
	fmt.Println("================================")
	demonstrateConcurrentFetch(ctx, tokenFetcher)

	// Example 3: Fetch with ETags for caching
	fmt.Println("\n💾 ETag-based Caching")
	fmt.Println("=====================")
	demonstrateETagCaching(ctx, tokenFetcher)

	// Example 4: Fetch list of token lists
	fmt.Println("\n📚 List of Token Lists")
	fmt.Println("======================")
	demonstrateListOfTokenLists(ctx, tokenFetcher)

	fmt.Println("\n✅ Token Fetcher examples completed!")
}

func demonstrateSingleFetch(ctx context.Context, tokenFetcher fetcher.Fetcher) {
	// Fetch Uniswap token list
	fetchDetails := fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "uniswap-default",
			SourceURL: "https://tokens.uniswap.org",
			Schema:    "",
		},
		Etag: "", // No ETag for first fetch
	}

	fmt.Printf("🔄 Fetching token list from: %s\n", fetchDetails.SourceURL)

	fetchedData, err := tokenFetcher.Fetch(ctx, fetchDetails)
	if err != nil {
		log.Printf("❌ Failed to fetch token list: %v", err)
		return
	}

	if len(fetchedData.JsonData) == 0 {
		fmt.Println("⚠️ No new data (possibly cached)")
		return
	}

	fmt.Printf("✅ Successfully fetched token list:\n")
	fmt.Printf("  📊 Data size: %d bytes\n", len(fetchedData.JsonData))
	fmt.Printf("  🏷️  ETag: %s\n", fetchedData.Etag)
	fmt.Printf("  📅 Fetched at: %s\n", fetchedData.Fetched.Format(time.RFC3339))

	// Show preview of data
	if len(fetchedData.JsonData) > 500 {
		fmt.Printf("  👀 Preview: %s...\n", string(fetchedData.JsonData[:500]))
	} else {
		fmt.Printf("  📄 Data: %s\n", string(fetchedData.JsonData))
	}
}

func demonstrateConcurrentFetch(ctx context.Context, tokenFetcher fetcher.Fetcher) {
	// Prepare multiple fetch requests
	fetchRequests := []fetcher.FetchDetails{
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-default",
				SourceURL: "https://tokens.uniswap.org",
				Schema:    "",
			},
		},
		{
			ListDetails: types.ListDetails{
				ID:        "compound-tokens",
				SourceURL: "https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
				Schema:    "",
			},
		},
		{
			ListDetails: types.ListDetails{
				ID:        "status-token-list",
				SourceURL: "https://prod.market.status.im/static/token-list.json",
				Schema:    "",
			},
		},
	}

	fmt.Printf("🚀 Fetching %d token lists concurrently...\n", len(fetchRequests))

	startTime := time.Now()
	results, err := tokenFetcher.FetchConcurrent(ctx, fetchRequests)
	if err != nil {
		log.Printf("❌ Concurrent fetch failed: %v", err)
		return
	}
	duration := time.Since(startTime)

	fmt.Printf("⚡ Concurrent fetch completed in %v\n\n", duration)

	// Process results
	successCount := 0
	for _, result := range results {
		fmt.Printf("📋 Token List: %s\n", result.ID)
		fmt.Printf("  🔗 URL: %s\n", result.SourceURL)

		if result.Error != nil {
			fmt.Printf("  ❌ Error: %v\n", result.Error)
		} else {
			fmt.Printf("  ✅ Success: %d bytes\n", len(result.JsonData))
			fmt.Printf("  🏷️  ETag: %s\n", result.Etag)
			fmt.Printf("  📅 Fetched: %s\n", result.Fetched.Format(time.RFC3339))
			successCount++
		}
		fmt.Println()
	}

	fmt.Printf("📊 Summary: %d/%d token lists fetched successfully\n",
		successCount, len(results))
}

func demonstrateETagCaching(ctx context.Context, tokenFetcher fetcher.Fetcher) {
	fetchDetails := fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "uniswap-default",
			SourceURL: "https://tokens.uniswap.org",
			Schema:    "",
		},
		Etag: "", // First fetch without ETag
	}

	fmt.Println("🔄 First fetch (no ETag)...")
	firstFetch, err := tokenFetcher.Fetch(ctx, fetchDetails)
	if err != nil {
		log.Printf("❌ First fetch failed: %v", err)
		return
	}

	if len(firstFetch.JsonData) > 0 {
		fmt.Printf("✅ First fetch successful: %d bytes, ETag: %s\n",
			len(firstFetch.JsonData), firstFetch.Etag)

		// Second fetch with ETag
		fmt.Println("\n🔄 Second fetch (with ETag)...")
		fetchDetails.Etag = firstFetch.Etag

		secondFetch, err := tokenFetcher.Fetch(ctx, fetchDetails)
		if err != nil {
			log.Printf("❌ Second fetch failed: %v", err)
			return
		}

		if len(secondFetch.JsonData) == 0 {
			fmt.Printf("💾 Cached response (304 Not Modified) - ETag: %s\n", secondFetch.Etag)
			fmt.Println("   No data transfer needed, content unchanged!")
		} else {
			fmt.Printf("📥 Content updated: %d bytes, New ETag: %s\n",
				len(secondFetch.JsonData), secondFetch.Etag)
		}
	} else {
		fmt.Println("⚠️ First fetch returned no data")
	}
}

func demonstrateListOfTokenLists(ctx context.Context, tokenFetcher fetcher.Fetcher) {
	// Fetch Status token lists
	fetchDetails := fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "status-lists",
			SourceURL: "https://prod.market.status.im/static/lists.json",
			Schema:    fetcher.ListOfTokenListsSchema,
		},
	}

	fmt.Printf("🔄 Fetching list of token lists from: %s\n", fetchDetails.SourceURL)

	fetchedData, err := tokenFetcher.Fetch(ctx, fetchDetails)
	if err != nil {
		log.Printf("❌ Failed to fetch list of token lists: %v", err)
		return
	}

	if len(fetchedData.JsonData) == 0 {
		fmt.Println("⚠️ No data received")
		return
	}

	fmt.Printf("✅ Successfully fetched list of token lists:\n")
	fmt.Printf("  📊 Data size: %d bytes\n", len(fetchedData.JsonData))
	fmt.Printf("  🏷️  ETag: %s\n", fetchedData.Etag)
	fmt.Printf("  📅 Fetched at: %s\n", fetchedData.Fetched.Format(time.RFC3339))

	// Show preview
	preview := string(fetchedData.JsonData)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Printf("  👀 Content preview:\n%s\n", preview)

	// Try to fetch first few token lists from the response
	fmt.Println("\n🔄 Attempting to fetch individual token lists...")

	// Note: In a real scenario, you would parse the JSON to extract URLs
	// For demo purposes, we'll show how you might process the result
	fmt.Println("  💡 Tip: Parse the JSON response to extract individual token list URLs")
	fmt.Println("      Then use FetchConcurrent() to fetch all lists in parallel")
}

// Example helper functions that might be used in a real application

func demonstrateErrorHandling() {
	fmt.Println("\n🛠️ Error Handling Examples")
	fmt.Println("==========================")

	tokenFetcher := fetcher.New(fetcher.DefaultConfig())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test cases for different error scenarios
	errorTests := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "Invalid URL",
			url:         "not-a-valid-url",
			expectedErr: "invalid URL",
		},
		{
			name:        "Unreachable Host",
			url:         "https://definitely-does-not-exist-12345.com/tokens.json",
			expectedErr: "network error",
		},
		{
			name:        "404 Not Found",
			url:         "https://httpbin.org/status/404",
			expectedErr: "HTTP 404",
		},
	}

	for _, test := range errorTests {
		fmt.Printf("\n🧪 Testing: %s\n", test.name)

		fetchDetails := fetcher.FetchDetails{
			ListDetails: types.ListDetails{
				ID:        "error-test",
				SourceURL: test.url,
				Schema:    "test",
			},
		}

		_, err := tokenFetcher.Fetch(ctx, fetchDetails)
		if err != nil {
			fmt.Printf("  ❌ Expected error occurred: %v\n", err)
		} else {
			fmt.Printf("  ⚠️ Unexpected success\n")
		}
	}
}

func demonstrateAdvancedUsage() {
	fmt.Println("\n🎯 Advanced Usage Patterns")
	fmt.Println("===========================")

	tokenFetcher := fetcher.New(fetcher.DefaultConfig())
	ctx := context.Background()

	// Pattern 1: Batch fetching with error tolerance
	fmt.Println("\n1️⃣ Batch fetching with error tolerance:")

	urls := []string{
		"https://tokens.uniswap.org",
		"https://raw.githubusercontent.com/compound-finance/token-list/master/compound.tokenlist.json",
		"https://invalid-url-that-will-fail.com/tokens.json", // This will fail
	}

	var fetchDetails []fetcher.FetchDetails
	for i, url := range urls {
		fetchDetails = append(fetchDetails, fetcher.FetchDetails{
			ListDetails: types.ListDetails{
				ID:        fmt.Sprintf("list-%d", i+1),
				SourceURL: url,
				Schema:    "",
			},
		})
	}

	results, err := tokenFetcher.FetchConcurrent(ctx, fetchDetails)
	if err != nil {
		fmt.Printf("  ❌ Batch fetch failed: %v\n", err)
	} else {
		successCount := 0
		for _, result := range results {
			if result.Error == nil {
				successCount++
			}
		}
		fmt.Printf("  ✅ Batch completed: %d/%d successful\n", successCount, len(results))
	}

	// Pattern 2: Retry logic with exponential backoff
	fmt.Println("\n2️⃣ Implementing retry logic:")
	fmt.Println("  💡 Tip: Wrap fetcher calls with retry logic:")
	fmt.Println("      - Exponential backoff for temporary failures")
	fmt.Println("      - Circuit breaker for persistent failures")
	fmt.Println("      - Timeout handling for slow responses")

	// Pattern 3: Caching strategy
	fmt.Println("\n3️⃣ Implementing caching strategy:")
	fmt.Println("  💡 Tip: Use ETags effectively:")
	fmt.Println("      - Store ETags with cached data")
	fmt.Println("      - Check for 304 Not Modified responses")
	fmt.Println("      - Implement cache expiration policies")
}

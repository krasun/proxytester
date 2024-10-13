package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	proxyURL := flag.String("proxy", "", "Specify the proxy URL you want to test")
	targetURL := flag.String("target", "https://example.com", "Specify the target URL used to test the proxy")
	times := flag.Int("requests", 1, "The number of requests to execute")
	failOnError := flag.Bool("fail-on-error", false, "Fail on the first encountered error")

	flag.Parse()

	if *proxyURL == "" {
		return fmt.Errorf("Proxy URL is required")
	}

	proxy, err := url.Parse(*proxyURL)
	if err != nil {
		fmt.Println("url.Parse:", err)
		return fmt.Errorf("url.Parse: %w", err)
	}

	target, err := url.Parse(*targetURL)
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	requestMetrics, err := testProxy(proxy, target, *times, *failOnError)
	if err != nil {
		return fmt.Errorf("testProxy: %w", err)
	}

	results := aggregateRequestMetrics(requestMetrics)

	fmt.Println("\nResults:")
	fmt.Printf("%-25s %-15s %-15s %-15s\n", "Metric", "Average", "P95", "Unit")
	fmt.Printf("%-25s %-15.2f %-15.2f %-15s\n", "Connect Time", results.AverageConnectTime.Seconds(), results.P95ConnectTime.Seconds(), "seconds")
	fmt.Printf("%-25s %-15.2f %-15.2f %-15s\n", "First Byte Time", results.AverageFirstByteTime.Seconds(), results.P95FirstByteTime.Seconds(), "seconds")
	fmt.Printf("%-25s %-15.2f %-15.2f %-15s\n", "Total Time", results.AverageTotalTime.Seconds(), results.P95TotalTime.Seconds(), "seconds")
	fmt.Printf("%-25s %-15d %-15s %-15s\n", "Error Count", results.ErrorCount, "-", "-")
	fmt.Printf("%-25s %-15.2f %-15s %-15s\n", "Error Rate", results.ErrorRate*100, "-", "%")

	fmt.Println("\nStatus Code Distribution:")
	fmt.Printf("%-15s %-15s\n", "Status Code", "Count")
	for code, count := range results.StatusCodes {
		fmt.Printf("%-15d %-15d\n", code, count)
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

type RequestMetrics struct {
	ConnectTime   time.Duration
	FirstByteTime time.Duration
	TotalTime     time.Duration
	StatusCode    int
	Error         error
}

type Results struct {
	AverageConnectTime   time.Duration
	AverageFirstByteTime time.Duration
	AverageTotalTime     time.Duration

	P95ConnectTime   time.Duration
	P95FirstByteTime time.Duration
	P95TotalTime     time.Duration

	StatusCodes map[int]int

	ErrorCount int
	ErrorRate  float64
}

func performRequest(proxyUrl *url.URL, targetUrl *url.URL) (*RequestMetrics, error) {
	httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	start := time.Now()
	requestMetrics := &RequestMetrics{}

	trace := &httptrace.ClientTrace{
		ConnectStart:         func(_, _ string) { requestMetrics.ConnectTime = time.Since(start) },
		GotFirstResponseByte: func() { requestMetrics.FirstByteTime = time.Since(start) },
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(context.Background(), trace), "GET", targetUrl.String(), nil)
	if err != nil {
		requestMetrics.Error = fmt.Errorf("http.NewRequestWithContext: %w", err)

		return requestMetrics, requestMetrics.Error
	}

	response, err := httpClient.Do(req)
	if err != nil {
		requestMetrics.Error = fmt.Errorf("httpClient.Do: %w", err)

		return requestMetrics, requestMetrics.Error
	}

	defer response.Body.Close()

	requestMetrics.TotalTime = time.Since(start)
	requestMetrics.StatusCode = response.StatusCode

	return requestMetrics, nil
}

func aggregateRequestMetrics(requestMetrics []*RequestMetrics) *Results {
	results := &Results{StatusCodes: make(map[int]int)}

	var totalConnectTime, totalFirstByteTime, totalResponseTime time.Duration
	var connectTimes, firstByteTimes, responseTimes []time.Duration
	errorCount := 0
	for _, result := range requestMetrics {
		if result.StatusCode != 0 {
			count, exists := results.StatusCodes[result.StatusCode]
			if exists {
				results.StatusCodes[result.StatusCode] = count + 1
			} else {
				results.StatusCodes[result.StatusCode] = 1
			}
		}

		if result.Error != nil {
			errorCount++
			continue
		}

		totalConnectTime += result.ConnectTime
		totalFirstByteTime += result.FirstByteTime
		totalResponseTime += result.TotalTime

		connectTimes = append(connectTimes, result.ConnectTime)
		firstByteTimes = append(firstByteTimes, result.FirstByteTime)
		responseTimes = append(responseTimes, result.TotalTime)
	}

	successfulRequests := len(requestMetrics) - errorCount
	if successfulRequests > 0 {
		results.AverageConnectTime = totalConnectTime / time.Duration(successfulRequests)
		results.AverageFirstByteTime = totalFirstByteTime / time.Duration(successfulRequests)
		results.AverageTotalTime = totalResponseTime / time.Duration(successfulRequests)

		results.P95ConnectTime = percentile(connectTimes, 95)
		results.P95FirstByteTime = percentile(firstByteTimes, 95)
		results.P95TotalTime = percentile(responseTimes, 95)
	}

	results.ErrorCount = errorCount
	results.ErrorRate = float64(errorCount) / float64(len(requestMetrics))

	return results
}

func testProxy(proxyUrl *url.URL, targetUrl *url.URL, times int, failOnError bool) ([]*RequestMetrics, error) {
	results := make([]*RequestMetrics, times)

	for t := 0; t < times; t++ {
		requestMetrics, err := performRequest(proxyUrl, targetUrl)
		results[t] = requestMetrics
		if err != nil && failOnError {
			return results, err
		}
	}

	return results, nil
}

func percentile(data []time.Duration, percentile float64) time.Duration {
	if len(data) == 0 {
		return 0
	}

	index := int(float64(len(data)) * percentile / 100)
	if index >= len(data) {
		index = len(data) - 1
	}

	return data[index]
}

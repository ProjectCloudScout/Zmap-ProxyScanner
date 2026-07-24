/*
	(c) Yariya
*/

package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	imported      uint64
	checked       uint64
	success       uint64
	statusCodeErr uint64

	proxyErr   uint64
	timeoutErr uint64
)

func formatSummary(importedCount, checkedCount, successCount, statusCodeErrCount, proxyErrCount, timeoutErrCount uint64, openThreads int64) string {
	return fmt.Sprintf("live | imported=%d checked=%d success=%d status_err=%d proxy_err=%d timeout=%d threads=%d",
		importedCount,
		checkedCount,
		successCount,
		statusCodeErrCount,
		proxyErrCount,
		timeoutErrCount,
		openThreads,
	)
}

func PrintSummary() {
	fmt.Printf("\nFinal summary: %s\n", formatSummary(
		atomic.LoadUint64(&imported),
		atomic.LoadUint64(&checked),
		atomic.LoadUint64(&success),
		atomic.LoadUint64(&statusCodeErr),
		atomic.LoadUint64(&proxyErr),
		atomic.LoadUint64(&timeoutErr),
		atomic.LoadInt64(&Proxies.openHttpThreads),
	))
}

func Stater() {
	for range time.Tick(time.Second) {
		fmt.Printf("%s\n", formatSummary(
			atomic.LoadUint64(&imported),
			atomic.LoadUint64(&checked),
			atomic.LoadUint64(&success),
			atomic.LoadUint64(&statusCodeErr),
			atomic.LoadUint64(&proxyErr),
			atomic.LoadUint64(&timeoutErr),
			atomic.LoadInt64(&Proxies.openHttpThreads),
		))
	}
}

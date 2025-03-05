package utils

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ResourceStats holds execution statistics
type ResourceStats struct {
    MemoryUsedBytes uint64
    ExecutionTime   time.Duration
    Error           error
}

// ResourceLimits defines the constraints for execution
type ResourceLimits struct {
    MaxMemoryBytes uint64
    MaxTimeMs      int
    MaxThreads     int
}

var DefaultResourceLimit = ResourceLimits{ MaxMemoryBytes: 60 * 1024,
    MaxTimeMs: 20,
    MaxThreads: 2,
}

// RunWithConstraints executes a function with resource constraints
func RunWithConstraints(ctx context.Context, limits ResourceLimits, fn func(ctx context.Context) error) ResourceStats {
    stats := ResourceStats{}
    
    // Create a context with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(limits.MaxTimeMs)*time.Millisecond)
    defer cancel()

    // Channel to collect results
    done := make(chan struct{})

    // Get initial memory stats
    var initialMemStats runtime.MemStats
    runtime.ReadMemStats(&initialMemStats)

    // Set max threads (GOMAXPROCS)
    originalMaxProcs := runtime.GOMAXPROCS(limits.MaxThreads)
    defer runtime.GOMAXPROCS(originalMaxProcs)

    startTime := time.Now()

    go func() {
        defer close(done)

        // Run the function
        stats.Error = fn(ctx)

        // Get final memory stats
        var finalMemStats runtime.MemStats
        runtime.ReadMemStats(&finalMemStats)

        // Calculate memory used (considering both heap and stack)
        stats.MemoryUsedBytes = finalMemStats.Alloc - initialMemStats.Alloc
        
        // Check if memory limit was exceeded
        if stats.MemoryUsedBytes > limits.MaxMemoryBytes {
            stats.Error = fmt.Errorf("memory limit exceeded: used %d bytes, limit %d bytes",
                stats.MemoryUsedBytes, limits.MaxMemoryBytes)
        }
    }()

    // Wait for either completion or timeout
    select {
    case <-timeoutCtx.Done():
        stats.Error = fmt.Errorf("execution timed out after %d ms", limits.MaxTimeMs)
    case <-done:
        // Function completed normally
    }

    stats.ExecutionTime = time.Since(startTime)
    return stats
}

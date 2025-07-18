package service

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryOptimizer manages memory usage during large import operations
type MemoryOptimizer struct {
	maxMemoryUsage     int64         // Maximum memory usage in bytes
	checkInterval      time.Duration // How often to check memory usage
	forceGCThreshold   float64       // Force GC when memory usage exceeds this percentage
	mu                 sync.RWMutex
	isMonitoring       bool
	stopChan          chan struct{}
	memoryStats       MemoryStats
	lastGCTime        time.Time
	gcCount           int
}

// MemoryStats tracks memory usage statistics
type MemoryStats struct {
	AllocatedMemory    int64     `json:"allocated_memory"`
	TotalAllocated     int64     `json:"total_allocated"`
	SystemMemory       int64     `json:"system_memory"`
	GCCount            int       `json:"gc_count"`
	LastGCTime         time.Time `json:"last_gc_time"`
	MaxMemoryUsage     int64     `json:"max_memory_usage"`
	CurrentMemoryUsage int64     `json:"current_memory_usage"`
	MemoryPressure     float64   `json:"memory_pressure"` // 0-1 scale
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(maxMemoryMB int64, checkInterval time.Duration) *MemoryOptimizer {
	if maxMemoryMB <= 0 {
		maxMemoryMB = 1024 // Default 1GB
	}
	if checkInterval <= 0 {
		checkInterval = 10 * time.Second
	}

	return &MemoryOptimizer{
		maxMemoryUsage:   maxMemoryMB * 1024 * 1024,
		checkInterval:    checkInterval,
		forceGCThreshold: 0.8, // Force GC at 80% of max memory
		stopChan:         make(chan struct{}),
	}
}

// StartMonitoring starts memory monitoring
func (m *MemoryOptimizer) StartMonitoring(ctx context.Context) {
	m.mu.Lock()
	if m.isMonitoring {
		m.mu.Unlock()
		return
	}
	m.isMonitoring = true
	m.mu.Unlock()

	go m.monitorMemory(ctx)
}

// StopMonitoring stops memory monitoring
func (m *MemoryOptimizer) StopMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isMonitoring {
		return
	}

	m.isMonitoring = false
	close(m.stopChan)
}

// monitorMemory continuously monitors memory usage
func (m *MemoryOptimizer) monitorMemory(ctx context.Context) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkAndOptimizeMemory()
		}
	}
}

// checkAndOptimizeMemory checks current memory usage and optimizes if needed
func (m *MemoryOptimizer) checkAndOptimizeMemory() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	m.memoryStats.AllocatedMemory = int64(memStats.Alloc)
	m.memoryStats.TotalAllocated = int64(memStats.TotalAlloc)
	m.memoryStats.SystemMemory = int64(memStats.Sys)
	m.memoryStats.CurrentMemoryUsage = int64(memStats.Alloc)
	m.memoryStats.MemoryPressure = float64(memStats.Alloc) / float64(m.maxMemoryUsage)

	// Track maximum memory usage
	if m.memoryStats.CurrentMemoryUsage > m.memoryStats.MaxMemoryUsage {
		m.memoryStats.MaxMemoryUsage = m.memoryStats.CurrentMemoryUsage
	}

	shouldForceGC := m.memoryStats.MemoryPressure >= m.forceGCThreshold
	m.mu.Unlock()

	// Force garbage collection if memory pressure is high
	if shouldForceGC {
		m.forceGarbageCollection()
	}
}

// forceGarbageCollection forces garbage collection
func (m *MemoryOptimizer) forceGarbageCollection() {
	m.mu.Lock()
	timeSinceLastGC := time.Since(m.lastGCTime)
	m.mu.Unlock()

	// Don't force GC too frequently (minimum 30 seconds between forced GCs)
	if timeSinceLastGC < 30*time.Second {
		return
	}

	beforeGC := m.getCurrentMemoryUsage()
	runtime.GC()
	afterGC := m.getCurrentMemoryUsage()

	m.mu.Lock()
	m.gcCount++
	m.lastGCTime = time.Now()
	m.memoryStats.GCCount = m.gcCount
	m.memoryStats.LastGCTime = m.lastGCTime
	m.mu.Unlock()

	fmt.Printf("Forced GC: memory reduced from %d MB to %d MB (saved %d MB)\n",
		beforeGC/1024/1024, afterGC/1024/1024, (beforeGC-afterGC)/1024/1024)
}

// getCurrentMemoryUsage returns current memory usage
func (m *MemoryOptimizer) getCurrentMemoryUsage() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc)
}

// GetMemoryStats returns current memory statistics
func (m *MemoryOptimizer) GetMemoryStats() MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryStats
}

// IsMemoryPressureHigh returns true if memory pressure is high
func (m *MemoryOptimizer) IsMemoryPressureHigh() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryStats.MemoryPressure >= 0.7 // 70% threshold
}

// GetMemoryPressure returns current memory pressure (0-1)
func (m *MemoryOptimizer) GetMemoryPressure() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryStats.MemoryPressure
}

// OptimizeForLargeImport optimizes settings for large import operations
func (m *MemoryOptimizer) OptimizeForLargeImport() {
	// Set lower GC threshold for large imports
	m.mu.Lock()
	m.forceGCThreshold = 0.6 // Force GC at 60% for large imports
	m.mu.Unlock()

	// Force an initial GC to start with a clean slate
	runtime.GC()
}

// RestoreNormalSettings restores normal memory optimization settings
func (m *MemoryOptimizer) RestoreNormalSettings() {
	m.mu.Lock()
	m.forceGCThreshold = 0.8 // Restore to 80%
	m.mu.Unlock()
}

// GetOptimalChunkSize returns optimal chunk size based on current memory usage
func (m *MemoryOptimizer) GetOptimalChunkSize(baseChunkSize int) int {
	pressure := m.GetMemoryPressure()
	
	// Reduce chunk size when memory pressure is high
	if pressure >= 0.8 {
		return baseChunkSize / 4 // Reduce to 25% of base size
	} else if pressure >= 0.6 {
		return baseChunkSize / 2 // Reduce to 50% of base size
	} else if pressure >= 0.4 {
		return int(float64(baseChunkSize) * 0.75) // Reduce to 75% of base size
	}
	
	return baseChunkSize
}

// ShouldPauseProcessing returns true if processing should be paused due to memory pressure
func (m *MemoryOptimizer) ShouldPauseProcessing() bool {
	return m.GetMemoryPressure() >= 0.9 // Pause at 90% memory usage
}

// WaitForMemoryAvailable waits until memory pressure drops below threshold
func (m *MemoryOptimizer) WaitForMemoryAvailable(ctx context.Context, maxWait time.Duration) error {
	if !m.ShouldPauseProcessing() {
		return nil
	}

	fmt.Println("Memory pressure high, waiting for memory to become available...")
	
	timeout := time.After(maxWait)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for memory to become available")
		case <-ticker.C:
			if !m.ShouldPauseProcessing() {
				fmt.Println("Memory pressure reduced, resuming processing...")
				return nil
			}
			// Force GC to help free memory
			runtime.GC()
		}
	}
}

// GetMemoryRecommendations returns recommendations for memory optimization
func (m *MemoryOptimizer) GetMemoryRecommendations() []string {
	stats := m.GetMemoryStats()
	var recommendations []string

	if stats.MemoryPressure >= 0.8 {
		recommendations = append(recommendations, "Memory pressure is high - consider reducing chunk size or processing fewer files concurrently")
	}

	if stats.MemoryPressure >= 0.6 {
		recommendations = append(recommendations, "Moderate memory pressure - enable streaming processing for better memory efficiency")
	}

	if stats.MaxMemoryUsage > m.maxMemoryUsage {
		recommendations = append(recommendations, fmt.Sprintf("Maximum memory usage exceeded limit - consider increasing memory limit or reducing batch size"))
	}

	if m.gcCount > 10 {
		recommendations = append(recommendations, "High GC frequency detected - consider optimizing data structures or reducing memory usage")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Memory usage is optimal")
	}

	return recommendations
}

// LogMemoryStats logs current memory statistics
func (m *MemoryOptimizer) LogMemoryStats() {
	stats := m.GetMemoryStats()
	fmt.Printf("Memory Stats: Allocated=%dMB, System=%dMB, Pressure=%.2f%%, GC Count=%d\n",
		stats.AllocatedMemory/1024/1024,
		stats.SystemMemory/1024/1024,
		stats.MemoryPressure*100,
		stats.GCCount)
}
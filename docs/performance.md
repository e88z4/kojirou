# Performance Optimization Guide

This document provides guidelines and strategies for optimizing the performance of Kojirou when generating e-books in different formats.

## Image Processing Optimization

### Reducing Memory Usage

Image processing is one of the most memory-intensive operations in Kojirou. To optimize memory usage:

1. **Process one image at a time**: Avoid loading all images into memory simultaneously
2. **Use streaming where possible**: Process images in streams rather than loading entire files
3. **Clean up temporary images**: Release memory after each image is processed
4. **Use appropriate image dimensions**: Scale images to target dimensions early in the pipeline

### Speeding Up Image Processing

1. **Parallel image processing**: Use goroutines for concurrent image processing
2. **Reuse image buffers**: Allocate buffers once and reuse them for multiple images
3. **Optimize image format conversion**: Use optimized libraries for format conversion
4. **Avoid repeated processing**: Process each image only once, then reuse the result

```go
// Example: Parallel image processing with worker pool
func processImagesInParallel(images []string, numWorkers int) []processedImage {
    var wg sync.WaitGroup
    results := make([]processedImage, len(images))
    jobs := make(chan imageJob, len(images))
    
    // Start workers
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results[job.index] = processImage(job.path)
            }
        }()
    }
    
    // Send jobs
    for i, path := range images {
        jobs <- imageJob{index: i, path: path}
    }
    close(jobs)
    
    wg.Wait()
    return results
}
```

## Format Generation Optimization

### Optimizing EPUB Generation

1. **Minimize DOM operations**: Batch DOM operations when creating EPUB HTML
2. **Use efficient templates**: Precompile templates for faster rendering
3. **Optimize CSS**: Keep stylesheets small and specific
4. **Batch file operations**: Minimize file system operations by batching writes

### Optimizing KEPUB Conversion

1. **Use incremental transformation**: Transform files incrementally rather than all at once
2. **Optimize HTML parsing**: Use efficient HTML parsing strategies
3. **Minimize regex use**: Replace regex with more efficient string operations where possible
4. **Stream ZIP operations**: Use streaming for ZIP file operations

## Memory Management

### Reducing Overall Memory Footprint

1. **Process volumes sequentially**: Process one volume at a time instead of entire series
2. **Use buffered I/O**: Use buffered I/O for large file operations
3. **Implement progressive loading**: Load and process content progressively
4. **Set appropriate buffer sizes**: Tune buffer sizes based on expected content size

```go
// Example: Processing volumes sequentially with memory cleanup
func processAllVolumes(manga Manga) error {
    for volID, volume := range manga.Volumes {
        // Process one volume
        err := processVolume(volume)
        if err != nil {
            return err
        }
        
        // Explicitly remove volume from memory after processing
        delete(manga.Volumes, volID)
        
        // Suggest garbage collection
        runtime.GC()
    }
    return nil
}
```

### Handling Large Manga Series

1. **Implement checkpoint system**: Save progress and resume for very large series
2. **Use disk caching**: Cache intermediate results to disk for large operations
3. **Implement low-memory mode**: Add an option for slower but memory-efficient processing
4. **Monitor memory usage**: Track memory usage and adjust behavior accordingly

## Parallel Processing

### Effective Use of Goroutines

1. **Balance parallelism**: Adjust number of goroutines based on available CPU cores
2. **Use worker pools**: Implement worker pools for controlled parallelism
3. **Avoid goroutine leaks**: Ensure all goroutines terminate properly
4. **Use appropriate synchronization**: Choose the right sync primitives for the task

### Format-Specific Parallelism

1. **Generate multiple formats in parallel**: Process different formats concurrently
2. **Process chapters in parallel**: Process different chapters simultaneously
3. **Parallel image operations**: Perform image operations concurrently
4. **Pipeline processing**: Implement pipeline pattern for stream processing

## Benchmarking and Profiling

### Tools for Performance Analysis

1. **Use Go's pprof**: Use built-in profiling tools to identify bottlenecks
2. **Measure performance**: Add benchmark tests for critical functions
3. **Profile memory usage**: Track memory allocation patterns
4. **Profile CPU usage**: Identify CPU-intensive operations

### Benchmarking Commands

```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench .

# Memory profiling
go test -memprofile mem.prof -bench .

# Block profiling
go test -blockprofile block.prof -bench .

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Implementation Recommendations

1. **Start with simplicity**: Begin with simple, clear implementations
2. **Measure before optimizing**: Profile to identify actual bottlenecks
3. **Optimize hotspots**: Focus optimization efforts on the most critical paths
4. **Balance readability and performance**: Maintain code readability while optimizing
5. **Document optimizations**: Explain performance-critical code sections

By following these guidelines, Kojirou can maintain excellent performance even when processing large manga series or generating multiple formats simultaneously.
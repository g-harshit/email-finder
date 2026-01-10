# Performance Optimization

## Latency Optimization

The service has been optimized to use **parallel email verification** instead of sequential processing.

### Performance Metrics

| Metric | Before (Sequential) | After (Parallel) | Improvement |
|--------|---------------------|------------------|-------------|
| **Latency (20 emails)** | ~28 seconds | ~3.4 seconds | **88% faster** |
| **Throughput** | 0.7 emails/sec | 5.9 emails/sec | **8.4x faster** |

### How It Works

1. **Parallel Processing**: Emails are verified concurrently using goroutines
2. **Concurrency Control**: Configurable worker pool (default: 10 concurrent verifications)
3. **Semaphore Pattern**: Prevents overwhelming the system with too many simultaneous requests
4. **Order Preservation**: Results maintain the same order as input emails

### Configuration

Set the concurrency level via environment variable:

```bash
VERIFICATION_CONCURRENCY=10  # Default: 10 concurrent verifications
```

**Recommended Settings:**
- **Low load**: 5-10 concurrent verifications
- **Medium load**: 10-20 concurrent verifications  
- **High load**: 20-50 concurrent verifications (adjust based on system resources)

### Technical Details

- Uses `sync.WaitGroup` for goroutine coordination
- Semaphore channel limits concurrent operations
- Each email verification runs in its own goroutine
- Results are collected in order using index mapping
- Error handling is per-email (one failure doesn't stop others)

### Example

**Before (Sequential):**
```
Email 1: 1.4s
Email 2: 1.4s
...
Email 20: 1.4s
Total: 28s
```

**After (Parallel with 10 workers):**
```
Emails 1-10: 1.4s (parallel)
Emails 11-20: 1.4s (parallel)
Total: ~2.8s + overhead = ~3.4s
```

### Monitoring

Check logs for latency metrics:
```bash
podman logs email-finder | grep latency
```

The service logs show the total request latency, which includes:
- Domain resolution
- Email pattern generation
- Parallel email verification
- Result processing

### Best Practices

1. **Adjust concurrency** based on your system's CPU and network capacity
2. **Monitor resource usage** - too high concurrency may cause timeouts
3. **Balance speed vs. reliability** - higher concurrency = faster but more resource usage
4. **Test with your workload** - optimal concurrency varies by use case

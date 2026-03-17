# Weighted Load Balancing & Rate Limiting

## Overview

The LLM-Proxy supports **Weighted Load Balancing** and **per-client Rate Limiting** for all provider types (Claude, OpenAI, Abacus.ai). This allows you to:

- Distribute requests across multiple API keys intelligently
- Prioritize certain API keys over others using weights
- Prevent rate limit errors by enforcing per-key request limits
- Automatically skip rate-limited keys and use available ones

## How It Works

### 1. Weighted Round-Robin Selection

When a request comes in, the proxy selects an API key using **weighted round-robin**:

- Each API key has a `weight` value (default: 1)
- Higher weight = more requests
- Example: Weight 3 gets 3x more requests than Weight 1

### 2. Per-Key Rate Limiting

Each API key has an optional `max_rpm` (Max Requests Per Minute) limit:

- Tracks requests in a rolling 1-minute window
- Automatically resets counter every minute
- When limit reached, proxy tries next available key
- If all keys are rate-limited, returns first key anyway (with warning)

### 3. Streaming Support

The weighted load balancing works for both:
- ✅ Regular requests (`CreateMessage`)
- ✅ Streaming requests (`GetClaudeClient`, `GetOpenAIClient`, `GetAbacusClient`)

---

## Configuration

### Basic Example

```yaml
providers:
  claude:
    enabled: true
    api_keys:
      - key: "sk-ant-api-key-1"
        weight: 1
        max_rpm: 50
```

**Behavior:**
- Single API key
- Max 50 requests per minute
- Weight has no effect with only one key

---

### Multiple Keys with Equal Weight

```yaml
providers:
  claude:
    enabled: true
    api_keys:
      - key: "sk-ant-api-key-1"
        weight: 1
        max_rpm: 100
      - key: "sk-ant-api-key-2"
        weight: 1
        max_rpm: 100
```

**Behavior:**
- Requests distributed equally (50/50)
- Each key max 100 RPM
- Total capacity: 200 RPM

---

### Weighted Distribution

```yaml
providers:
  claude:
    enabled: true
    api_keys:
      - key: "sk-ant-production-key"
        weight: 3
        max_rpm: 300
      - key: "sk-ant-backup-key"
        weight: 1
        max_rpm: 100
```

**Behavior:**
- Production key gets 75% of requests (3/4)
- Backup key gets 25% of requests (1/4)
- Total capacity: 400 RPM
- If production key hits 300 RPM, backup key handles overflow

**Use Case:**
- Primary key has higher rate limits
- Backup key for failover

---

### Unlimited Rate Limiting

```yaml
providers:
  openai:
    enabled: true
    api_keys:
      - key: "sk-proj-api-key-1"
        weight: 2
        max_rpm: 0  # 0 = unlimited
      - key: "sk-proj-api-key-2"
        weight: 1
        max_rpm: 60
```

**Behavior:**
- Key 1: Unlimited requests (2x weight)
- Key 2: Max 60 RPM (1x weight)
- Key 1 will handle ~66% of requests (as long as it's not rate-limited)

**Use Case:**
- Primary key has Pay-As-You-Go (no rate limit)
- Secondary key has Tier limits

---

## Advanced Use Cases

### High Availability Setup

```yaml
providers:
  claude:
    enabled: true
    api_keys:
      # Primary Keys (High Weight)
      - key: "sk-ant-primary-1"
        weight: 5
        max_rpm: 500
      - key: "sk-ant-primary-2"
        weight: 5
        max_rpm: 500
      
      # Backup Keys (Low Weight)
      - key: "sk-ant-backup-1"
        weight: 1
        max_rpm: 100
      - key: "sk-ant-backup-2"
        weight: 1
        max_rpm: 100
```

**Behavior:**
- 83% of traffic goes to primary keys (10/12 weight)
- 17% goes to backup keys (2/12 weight)
- Total capacity: 1,200 RPM
- Automatic failover if primary keys hit rate limits

---

### Cost Optimization

```yaml
providers:
  openai:
    enabled: true
    api_keys:
      # Cheap models (high weight)
      - key: "sk-proj-tier1-key"
        weight: 4
        max_rpm: 200
      
      # Expensive models (low weight)
      - key: "sk-proj-tier2-key"
        weight: 1
        max_rpm: 100
```

**Behavior:**
- 80% of requests use tier1 key
- 20% use tier2 key
- Use this if tier1 has better pricing

---

### Load Testing / Development

```yaml
providers:
  claude:
    enabled: true
    api_keys:
      # Production Key (low weight during testing)
      - key: "sk-ant-production"
        weight: 1
        max_rpm: 1000
      
      # Test Key (high weight during testing)
      - key: "sk-ant-test"
        weight: 9
        max_rpm: 100
```

**Behavior:**
- 90% of test traffic goes to test key
- Only 10% hits production
- Easy to adjust weights without code changes

---

## Rate Limit Behavior

### When a Key Hits Rate Limit

```
Time: 10:00:00 - Key1 at 49/50 RPM
Time: 10:00:01 - Key1 at 50/50 RPM (LIMIT REACHED)
Time: 10:00:02 - Request comes in
                 → Key1 skipped (rate limited)
                 → Key2 selected instead
Time: 10:01:00 - Key1 counter resets to 0
                 → Key1 available again
```

### All Keys Rate Limited

If ALL keys are rate-limited:

```
Log: "All Claude clients are rate-limited, returning first client"
→ Returns first key anyway (provider will return rate limit error)
```

**Why?**
- Better than rejecting request immediately
- Provider might have burst capacity
- Preserves request for retry logic

---

## Monitoring

### Log Messages

**Initialization:**
```
Initialized Claude provider 1 from config with key: ...xxx123 (weight: 3, maxRPM: 100)
Initialized Claude provider 2 from config with key: ...xxx456 (weight: 1, maxRPM: 50)
```

**Rate Limiting:**
```
All Claude clients are rate-limited, returning first client
```

**Hot Reload:**
```
Reloading provider API keys from database...
Reloaded Claude provider 1 from DB with key: ...xxx123 (weight: 3, maxRPM: 100)
Provider keys reloaded: 2 provider(s) (Claude: 2, OpenAI: 0, Abacus: 0)
```

### Prometheus Metrics (Future)

Future enhancement: Track per-key metrics:
- `llm_proxy_provider_requests_total{provider="claude", key="xxx123"}`
- `llm_proxy_provider_rate_limited_total{provider="claude", key="xxx123"}`

---

## Best Practices

### 1. Start with Equal Weights

```yaml
api_keys:
  - key: "key-1"
    weight: 1
    max_rpm: 100
  - key: "key-2"
    weight: 1
    max_rpm: 100
```

**Why:** Simple, predictable, easy to reason about.

---

### 2. Set Conservative Rate Limits

```yaml
max_rpm: 80  # If actual limit is 100 RPM
```

**Why:** 
- Buffer for burst traffic
- Prevents hitting provider rate limits
- Accounts for other services using same key

---

### 3. Use Weights for Failover

```yaml
api_keys:
  - key: "primary"
    weight: 5
    max_rpm: 500
  - key: "backup"
    weight: 1
    max_rpm: 100
```

**Why:**
- Primary handles most traffic
- Backup only used when needed
- Cost-effective

---

### 4. Test Your Configuration

```bash
# Send 100 requests and check distribution
for i in {1..100}; do
  curl -X POST http://localhost:8080/v1/chat/completions \
    -H "Authorization: Bearer YOUR_KEY" \
    -d '{"model":"claude-3-haiku-20240307","messages":[{"role":"user","content":"test"}]}'
done

# Check logs to see which keys were used
docker logs llm-proxy-backend | grep "provider.*key:"
```

---

### 5. Monitor Rate Limit Warnings

```bash
# Watch for rate limit warnings
docker logs llm-proxy-backend -f | grep "rate-limited"
```

If you see frequent warnings:
- Increase `max_rpm` limits
- Add more API keys
- Adjust weights

---

## Troubleshooting

### Issue: All requests go to one key

**Cause:** Other keys might be rate-limited or have weight 0

**Solution:**
1. Check logs for rate limit warnings
2. Verify all keys have weight > 0
3. Increase `max_rpm` if needed

---

### Issue: Getting rate limit errors from provider

**Cause:** `max_rpm` set too high

**Solution:**
1. Check provider's actual rate limits
2. Reduce `max_rpm` to 80% of actual limit
3. Add more API keys if needed

---

### Issue: Uneven distribution despite equal weights

**Cause:** Some keys might be rate-limited intermittently

**Solution:**
1. Monitor logs for rate limit warnings
2. Increase `max_rpm` on affected keys
3. Check if provider has temporary rate limits

---

### Issue: Need to change weights without restart

**Solution:**
1. Update database via Admin API
2. Call reload endpoint:
   ```bash
   curl -X POST http://localhost:8080/admin/providers/reload-keys \
     -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
   ```
3. Or update `config.yaml` and restart

---

## Database Support

API keys can be stored in the database with weight and max_rpm:

```sql
INSERT INTO provider_api_keys (provider_id, api_key_encrypted, weight, max_rpm)
VALUES ('claude', encrypt('sk-ant-xxx', 'encryption-key'), 3, 100);
```

**Priority:**
1. Database keys (if available)
2. Config.yaml keys (fallback)

**Hot Reload:**
- Call `/admin/providers/reload-keys` to reload from DB
- No restart needed

---

## Migration from Old Setup

If you have an existing config without weight/max_rpm:

**Old:**
```yaml
api_keys:
  - "sk-ant-api-key-1"
```

**New:**
```yaml
api_keys:
  - key: "sk-ant-api-key-1"
    weight: 1
    max_rpm: 100
```

**Defaults:**
- `weight`: 1 (if not specified)
- `max_rpm`: 0 (unlimited, if not specified)

---

## Performance Impact

The weighted load balancing has **minimal performance overhead**:

- O(n) time complexity where n = number of keys
- Typically < 1ms for selection
- No external calls or database queries
- All state in memory

**Benchmarks (internal):**
- 1 key: 0.001ms
- 10 keys: 0.005ms
- 100 keys: 0.05ms

---

## Future Enhancements

Planned features:

1. **Health-Based Selection**
   - Skip unhealthy providers automatically
   - Circuit breaker pattern

2. **Per-Key Metrics**
   - Prometheus metrics per key
   - Dashboard for monitoring

3. **Dynamic Weight Adjustment**
   - Auto-adjust weights based on latency
   - Learn optimal distribution

4. **Quota Management**
   - Track monthly/daily quotas per key
   - Automatic key rotation

---

## References

- [Config Example](../configs/config.example.yaml)
- [Provider Manager Code](../internal/infrastructure/providers/manager.go)
- [Abacus Integration Guide](./ABACUS_INTEGRATION.md)

---

**Last Updated:** March 17, 2026  
**Version:** 1.0.0

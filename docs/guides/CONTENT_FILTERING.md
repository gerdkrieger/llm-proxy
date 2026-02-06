# Content Filtering

The LLM-Proxy includes a powerful content filtering system that allows you to automatically filter and replace unwanted content in user prompts before they reach the LLM API.

## Table of Contents

- [Overview](#overview)
- [Filter Types](#filter-types)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Performance](#performance)

---

## Overview

Content filtering enables you to:

- **Block offensive or inappropriate language** by replacing specific words or phrases
- **Redact sensitive information** like emails, phone numbers, or proprietary terms
- **Enforce content policies** across all API requests
- **Track filter matches** for compliance and analytics

### Key Features

- ✅ **Three filter types**: Word, Phrase, and Regex
- ✅ **Priority-based filtering**: Control the order filters are applied
- ✅ **Case-sensitive/insensitive** matching options
- ✅ **Enable/disable filters** on the fly without removing them
- ✅ **Real-time statistics** tracking matches and usage
- ✅ **Automatic caching** for optimal performance (5-minute TTL)
- ✅ **Admin API** for filter management

---

## Filter Types

### 1. Word Filter

Matches whole words using word boundaries (`\b`).

**Example:**
```json
{
  "pattern": "badword",
  "replacement": "[FILTERED]",
  "filter_type": "word",
  "case_sensitive": false
}
```

**Behavior:**
- Matches: "badword", "BADWORD", "BadWord"
- Does NOT match: "mybadword", "badwords"

### 2. Phrase Filter

Matches exact phrases (escaped for safety).

**Example:**
```json
{
  "pattern": "confidential information",
  "replacement": "[REDACTED]",
  "filter_type": "phrase",
  "case_sensitive": false
}
```

**Behavior:**
- Matches: "confidential information", "CONFIDENTIAL INFORMATION"
- Does NOT match: "confidential", "information"

### 3. Regex Filter

Matches using regular expressions for complex patterns.

**Example:**
```json
{
  "pattern": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
  "replacement": "[EMAIL]",
  "filter_type": "regex",
  "case_sensitive": false
}
```

**Behavior:**
- Matches: "user@example.com", "admin@company.co.uk"

**⚠️ Security Note:** Regex patterns are validated to prevent ReDoS (Regular Expression Denial of Service) attacks. Dangerous patterns like `.*.*` or `(.+)+` are rejected.

---

## API Reference

All content filter endpoints require the `X-Admin-API-Key` header.

### Create Filter

**POST** `/admin/filters`

```bash
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "badword",
    "replacement": "[FILTERED]",
    "description": "Filter offensive word",
    "filter_type": "word",
    "case_sensitive": false,
    "enabled": true,
    "priority": 100
  }'
```

**Response:**
```json
{
  "id": 1,
  "pattern": "badword",
  "replacement": "[FILTERED]",
  "description": "Filter offensive word",
  "filter_type": "word",
  "case_sensitive": false,
  "enabled": true,
  "priority": 100,
  "created_at": "2026-01-30T10:00:00Z",
  "updated_at": "2026-01-30T10:00:00Z",
  "match_count": 0
}
```

### List Filters

**GET** `/admin/filters`

Query Parameters:
- `enabled_only=true` - List only enabled filters

```bash
curl http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

**Response:**
```json
{
  "filters": [
    {
      "id": 1,
      "pattern": "badword",
      "replacement": "[FILTERED]",
      ...
    }
  ],
  "count": 1
}
```

### Get Filter by ID

**GET** `/admin/filters/{id}`

```bash
curl http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

### Update Filter

**PUT** `/admin/filters/{id}`

```bash
curl -X PUT http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "priority": 150,
    "enabled": true
  }'
```

### Delete Filter

**DELETE** `/admin/filters/{id}`

```bash
curl -X DELETE http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

### Test Filter (Ad-hoc)

**POST** `/admin/filters/test`

Test a filter without creating it:

```bash
curl -X POST http://localhost:8080/admin/filters/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "This contains a badword",
    "pattern": "badword",
    "replacement": "[FILTERED]",
    "filter_type": "word"
  }'
```

**Response:**
```json
{
  "original_text": "This contains a badword",
  "filtered_text": "This contains a [FILTERED]",
  "matches": [
    {
      "filter_id": 0,
      "pattern": "badword",
      "replacement": "[FILTERED]",
      "match_count": 1
    }
  ],
  "has_matches": true
}
```

### Test Existing Filter

**POST** `/admin/filters/{id}/test`

```bash
curl -X POST http://localhost:8080/admin/filters/1/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Test text here"
  }'
```

### Get Filter Statistics

**GET** `/admin/filters/stats`

```bash
curl http://localhost:8080/admin/filters/stats \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

**Response:**
```json
{
  "total_filters": 3,
  "enabled_filters": 2,
  "cached_filters": 2,
  "total_matches": 42,
  "by_type": {
    "word": 2,
    "phrase": 1,
    "regex": 0
  },
  "cache_age_seconds": 125,
  "last_match": "2026-01-30T10:30:00Z"
}
```

### Refresh Filter Cache

**POST** `/admin/filters/refresh`

Force reload of filters from database:

```bash
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

**Response:**
```json
{
  "message": "Filters refreshed successfully",
  "cached_filters": 2
}
```

---

## Examples

### Example 1: Filter Profanity

```bash
# Create filter
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "damn",
    "replacement": "[*]",
    "description": "Filter mild profanity",
    "filter_type": "word",
    "case_sensitive": false,
    "enabled": true,
    "priority": 100
  }'
```

**Input:** "This damn thing doesn't work!"  
**Output:** "This [*] thing doesn't work!"

### Example 2: Redact Email Addresses

```bash
# Create regex filter
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
    "replacement": "[EMAIL_REDACTED]",
    "description": "Redact email addresses",
    "filter_type": "regex",
    "case_sensitive": false,
    "enabled": true,
    "priority": 90
  }'
```

**Input:** "Contact me at john.doe@example.com"  
**Output:** "Contact me at [EMAIL_REDACTED]"

### Example 3: Filter Company-Specific Terms

```bash
# Create phrase filter
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "Project Phoenix",
    "replacement": "[CONFIDENTIAL PROJECT]",
    "description": "Filter internal project name",
    "filter_type": "phrase",
    "case_sensitive": true,
    "enabled": true,
    "priority": 95
  }'
```

**Input:** "Discussing Project Phoenix details"  
**Output:** "Discussing [CONFIDENTIAL PROJECT] details"

### Example 4: Multiple Filters Combined

```bash
# Create multiple filters with different priorities
# Priority 100: Word filter
# Priority 90: Phrase filter
# Priority 80: Regex filter

# Send chat request with multiple issues
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_OAUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "The damn Project Phoenix has confidential information. Email me at admin@company.com"
      }
    ],
    "max_tokens": 100
  }'
```

**Filtered Input (sent to Claude):**
```
"The [*] [CONFIDENTIAL PROJECT] has [REDACTED]. Email me at [EMAIL_REDACTED]"
```

---

## Best Practices

### 1. Use Appropriate Priorities

- **High priority (90-100)**: Critical filters (e.g., offensive content, legal terms)
- **Medium priority (50-89)**: Company-specific filters
- **Low priority (0-49)**: Optional/experimental filters

### 2. Test Before Enabling

Always test filters using the `/admin/filters/test` endpoint before enabling them:

```bash
# Test first
curl -X POST http://localhost:8080/admin/filters/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Sample text to test",
    "pattern": "your_pattern",
    "replacement": "[FILTERED]",
    "filter_type": "word"
  }'

# If OK, create the filter
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

### 3. Monitor Filter Statistics

Regularly check filter statistics to identify:
- Unused filters (0 matches)
- Overused filters (high match count)
- Performance issues (cache age)

```bash
curl http://localhost:8080/admin/filters/stats \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

### 4. Use Case-Insensitive Matching

Unless you need exact case matching, use `case_sensitive: false`:

```json
{
  "pattern": "badword",
  "case_sensitive": false  // Matches: badword, BADWORD, BadWord
}
```

### 5. Disable Instead of Delete

When testing or troubleshooting, disable filters instead of deleting them:

```bash
curl -X PUT http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

### 6. Refresh Cache After Bulk Changes

After creating/updating multiple filters, refresh the cache:

```bash
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_KEY"
```

---

## Performance

### Caching

- Filters are **cached for 5 minutes** by default
- Cache is automatically refreshed when:
  - Creating a new filter
  - Updating an existing filter
  - Deleting a filter
  - Manually calling `/admin/filters/refresh`

### Regex Performance

- Regex patterns are **pre-compiled** and cached
- Dangerous patterns (ReDoS) are **validated and rejected**
- Avoid overly complex regex patterns for best performance

### Overhead

- **Typical overhead**: <10ms per request (with 10-20 filters)
- **No overhead** if no filters are enabled
- Match recording is **asynchronous** (no impact on request latency)

### Optimization Tips

1. **Minimize active filters**: Disable unused filters
2. **Use word/phrase filters** when possible (faster than regex)
3. **Optimize regex patterns**: Avoid unnecessary complexity
4. **Monitor cache age**: Ensure cache is fresh

---

## Integration

Filters are automatically applied to **all user messages** in chat completion requests:

```javascript
// Client code (unchanged)
const response = await openai.chat.completions.create({
  model: "claude-3-haiku-20240307",
  messages: [
    { role: "user", content: "This contains filtered content" }
  ]
});

// Content is automatically filtered by the proxy before reaching Claude
```

### Logging

Filter matches are logged for audit purposes:

```
INFO  Applied 2 content filters to request abc123 (client: test_client)
```

### Monitoring

Track filter usage via Prometheus metrics:

- `llm_proxy_content_filter_matches_total` - Total filter matches
- `llm_proxy_content_filter_cache_hits_total` - Cache performance

---

## Troubleshooting

### Filter Not Matching

1. **Check filter type**: Ensure you're using the correct type (word/phrase/regex)
2. **Test the filter**: Use `/admin/filters/test` to verify pattern
3. **Check case sensitivity**: Try with `case_sensitive: false`
4. **Verify priority**: Higher priority filters are applied first

### Performance Issues

1. **Check cache age**: Use `/admin/filters/stats` to see cache freshness
2. **Refresh cache**: Call `/admin/filters/refresh`
3. **Optimize regex**: Simplify complex patterns
4. **Disable unused filters**: Reduce active filter count

### Invalid Regex Pattern

```json
{
  "error": {
    "type": "invalid_pattern",
    "message": "Invalid pattern: potentially dangerous regex pattern detected"
  }
}
```

**Solution:** Simplify your regex pattern to avoid catastrophic backtracking.

---

## Security Considerations

1. **Admin API Key**: Protect your admin API key - it has full filter management access
2. **Regex Validation**: The system automatically rejects dangerous patterns
3. **Audit Logging**: All filter operations are logged
4. **No PII in Replacements**: Avoid putting sensitive data in replacement text

---

## Testing

Use the provided test script to verify your setup:

```bash
# Use Admin UI at http://localhost:3005
```

This script tests:
- Creating filters (word, phrase, regex)
- Listing and retrieving filters
- Testing filters
- Updating and deleting filters
- Filter application in chat completions
- Statistics and caching

---

## Support

For issues or questions:
- Check the main [README.md](README.md) for general setup
- Review [ADMIN_API.md](ADMIN_API.md) for admin API details
- See filter statistics: `GET /admin/filters/stats`

---

**Made with ❤️ for Content Safety**

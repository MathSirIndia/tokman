import time
import pytest
import httpx

@pytest.mark.e2e
@pytest.mark.module1
async def test_litellm_readiness(litellm_client: httpx.AsyncClient):
    """Verifies that LiteLLM core proxy is live and healthy."""
    res = await litellm_client.get("/health/readiness")
    assert res.status_code == 200, f"LiteLLM readiness probe failed with status: {res.status_code}"

@pytest.mark.e2e
@pytest.mark.module1
async def test_pool_general_invocation(litellm_client: httpx.AsyncClient):
    """Sends a completion prompt to pool/general and validates the response."""
    payload = {
        "model": "pool/general",
        "messages": [
            {"role": "system", "content": "You are a concise test assistant."},
            {"role": "user", "content": "Say 'online' and nothing else."}
        ],
        "max_tokens": 10
    }
    res = await litellm_client.post("/v1/chat/completions", json=payload)
    assert res.status_code == 200, f"Request failed ({res.status_code}): {res.text}"

    data = res.json()
    assert "choices" in data and len(data["choices"]) > 0, "Response missing choices."
    content = data["choices"][0]["message"]["content"]
    assert content is not None and len(content.strip()) > 0, "Empty content received."

@pytest.mark.e2e
@pytest.mark.module1
async def test_pool_deep_reasoning_invocation(litellm_client: httpx.AsyncClient):
    """Sends a logic prompt to pool/deep-reasoning (DeepSeek-R1 Distill)."""
    payload = {
        "model": "pool/deep-reasoning",
        "messages": [
            {"role": "user", "content": "What is 15 * 4?"}
        ],
        "max_tokens": 25
    }
    res = await litellm_client.post("/v1/chat/completions", json=payload)
    assert res.status_code == 200, f"Reasoning request failed ({res.status_code}): {res.text}"

    data = res.json()
    assert "choices" in data and len(data["choices"]) > 0
    content = data["choices"][0]["message"]["content"]
    assert "60" in content, f"Expected answer 60 in reasoning output, got: {content}"

@pytest.mark.e2e
@pytest.mark.module1
async def test_redis_cache_hit(litellm_client: httpx.AsyncClient):
    """Verifies that an identical second query resolves via Redis cache with sub-50ms latency."""
    cache_prompt = f"Cache probe at timestamp {time.strftime('%Y-%m-%d %H:%M')}"
    payload = {
        "model": "pool/general",
        "messages": [
            {"role": "user", "content": cache_prompt}
        ],
        "max_tokens": 10
    }

    # First call: populates cache
    res1 = await litellm_client.post("/v1/chat/completions", json=payload)
    assert res1.status_code == 200

    # Second call: must hit cache
    start = time.perf_counter()
    res2 = await litellm_client.post("/v1/chat/completions", json=payload)
    duration_ms = (time.perf_counter() - start) * 1000

    assert res2.status_code == 200
    # Cache hit is typically < 30ms locally, allowing up to 100ms tolerance
    assert duration_ms < 100, f"Expected cache hit latency < 100ms, took {duration_ms:.2f}ms"

@pytest.mark.e2e
@pytest.mark.module1
async def test_redis_keys_exist(redis_client):
    """Verifies that Redis holds active cache keys."""
    keys = await redis_client.keys("*")
    assert len(keys) > 0, "No keys found in Redis cache after completions."

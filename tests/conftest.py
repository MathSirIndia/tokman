import os
import pytest
import pytest_asyncio
import httpx
import redis.asyncio as aioredis
from pathlib import Path

# Resolve project root and load environment
PROJECT_ROOT = Path(__file__).parent.parent
ENV_PATH = PROJECT_ROOT / ".env"

if ENV_PATH.exists():
    with open(ENV_PATH, "r") as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#") and "=" in line:
                key, val = line.split("=", 1)
                os.environ.setdefault(key.strip(), val.strip())

LITELLM_URL = os.getenv("LITELLM_URL", "http://127.0.0.1:4000")
MASTER_KEY = os.getenv("LITELLM_MASTER_KEY", "sk-master-internal-network-key")
REDIS_HOST = os.getenv("REDIS_HOST", "127.0.0.1")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))

@pytest_asyncio.fixture(scope="function")
async def litellm_client():
    headers = {
        "Authorization": f"Bearer {MASTER_KEY}",
        "Content-Type": "application/json"
    }
    async with httpx.AsyncClient(base_url=LITELLM_URL, headers=headers, timeout=45.0) as client:
        yield client

@pytest_asyncio.fixture(scope="function")
async def redis_client():
    client = await aioredis.from_url(f"redis://{REDIS_HOST}:{REDIS_PORT}", decode_responses=True)
    try:
        yield client
    finally:
        await client.close()

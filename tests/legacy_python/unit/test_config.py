import pytest
import yaml
from pathlib import Path

CONFIG_PATH = Path(__file__).parent.parent.parent / "config" / "config.yaml"

@pytest.mark.unit
@pytest.mark.module1
def test_config_file_exists():
    assert CONFIG_PATH.exists(), f"Configuration file {CONFIG_PATH} does not exist."

@pytest.mark.unit
@pytest.mark.module1
def test_config_syntax_and_schema():
    with open(CONFIG_PATH, "r") as f:
        data = yaml.safe_load(f)

    assert isinstance(data, dict), "config.yaml must parse to a dictionary."
    assert "model_list" in data, "config.yaml must declare 'model_list'."
    assert len(data["model_list"]) >= 2, "Module 1 config must declare at least 2 models."

    pool_names = {entry.get("model_name") for entry in data["model_list"]}
    assert "pool/general" in pool_names, "Missing 'pool/general' capability pool."
    assert "pool/deep-reasoning" in pool_names, "Missing 'pool/deep-reasoning' capability pool."

    # Validate individual model parameters
    for entry in data["model_list"]:
        params = entry.get("litellm_params", {})
        assert "model" in params, f"Model definition {entry} missing 'model' identifier."
        assert "api_key" in params, f"Model definition {entry} missing 'api_key' parameter."

    # Router & cache settings
    assert "router_settings" in data, "Missing 'router_settings'."
    assert data["router_settings"].get("enable_pre_call_checks") is True, "Pre-call headroom checks must be enabled."

    assert "litellm_settings" in data, "Missing 'litellm_settings'."
    assert data["litellm_settings"].get("cache") is True, "Redis caching must be enabled."
    assert data["litellm_settings"].get("cache_type") == "redis", "Cache type must be 'redis'."

# Legacy Python Test Infrastructure (Archived)

These Python test files (`conftest.py`, `pytest.ini`, `requirements-test.txt`, `test_config.py`, `test_module1_core.py`) represent the pre-migration research test suite when the gateway was initially conceptualized with Python/FastAPI.

Following the architectural pivot to a pure native Go binary (`tokman`), all active unit and end-to-end tests are implemented in Go and executed via:
```bash
./tests/run_tests.sh --all
```
These legacy files are retained strictly for historical reference.

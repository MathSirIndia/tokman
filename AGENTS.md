# Antigravity Agent Guidelines: TokMan - Ultimate AI Orchestration

These rules govern the development workflow, testing standards, and tracking protocols for the **TokMan - Ultimate AI Orchestration** (`tokman`) repository.

---

## 1. Mandatory Progress Tracking Rule

> **Rule:** Immediately update [docs/progress.md](file:///home/gamers/dev/tokman/docs/progress.md) after completing any implementation task, script creation, configuration change, or verification step.

- **Checklist Sync:** Mark completed deliverables with `[x]` as soon as the file or feature is created and verified.
- **Metric Updates:** Adjust the **Global Progress %** and **Current State** in the Executive Dashboard to reflect real repository state.
- **Verification Log:** Record every test execution (date, module, test description, outcome, and notes) in the **Manual Verification Runbook Log** table.
- **Never proceed to a subsequent module** without updating `docs/progress.md`.

---

## 2. Automated Testing Enforcement Rule

> **Rule:** Execute the automated test suite after implementing each module and add new tests alongside every code change.

- **Pre-Execution Unit Testing:** Run offline unit tests (`pytest tests/unit/`) to validate pure algorithms, regexes, and mathematical logic before deploying or starting Docker containers.
- **Module Sanity Gate:** After implementing any module, execute the module-specific test suite:
  ```bash
  ./tests/run_tests.sh --module <MODULE_NUMBER>
  ```
- **Incremental Test Additions:**
  - Whenever new functionality or routes are added, write matching unit tests in `tests/unit/` and integration/e2e tests in `tests/e2e/`.
  - Edge cases (rate limiting, 429 retries, circuit breakers) must be accompanied by resiliency tests in `tests/resiliency/`.
- **Zero Regression Policy:** All existing automated tests from previous modules must continue to pass before concluding the current module.

---

## 3. Modular Development & Verification Gating

> **Rule:** Follow strict module-by-module isolation. Do not build future module components until the active module has passed automated tests and user verification.

- Adhere to the architectural definitions in [docs/roadmap.md](file:///home/gamers/dev/tokman/docs/roadmap.md) and the detailed specifications in [docs/modules/](file:///home/gamers/dev/tokman/docs/modules/).
- Preserve the directory layout defined in the Master Repository Directory Structure.

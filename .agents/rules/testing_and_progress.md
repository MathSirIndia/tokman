# Rule: Mandatory Progress Tracking & Automated Testing Protocol

## Trigger
Always active across all tasks in the `tokman` workspace.

## Instructions

1. **Progress Tracking Protocol (`docs/progress.md`)**:
   - Immediately update `docs/progress.md` after every file edit, script generation, or configuration change.
   - Check off completed items in the module's granular checklist.
   - Record test executions in the verification runbook log.

2. **Automated Testing Protocol (`tests/`)**:
   - For every module implemented, add unit tests in `tests/unit/` and e2e smoke tests in `tests/e2e/`.
   - Run the automated test suite using `./tests/run_tests.sh --module <N>` upon completing each module.
   - Ensure all tests pass with 100% assertions before moving to the next module.
   - Run regression tests to verify that changes have not broken existing functionality.

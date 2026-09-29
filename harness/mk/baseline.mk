# harness/mk/baseline.mk — Baseline-Fragment, emittiert von ai-harness-init.
# Verifiziert die vendored Baseline netzlos und haengt baseline-verify an GATE_CHECKS;
# der Root-Aggregator faehrt es via make gates.
.PHONY: baseline-verify

baseline-verify: ## Vendored Baseline netzlos verifizieren
	@bash tools/harness/baseline-verify.sh

GATE_CHECKS += baseline-verify

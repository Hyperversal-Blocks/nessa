# Non-interactive - Encrypted - Self-Sovereign - Session-Auth - Attestation (NESSA)

## Executive Overview
Project **NESSA** is a privacy-first, local-first cryptographic substrate for managing identity, access, and credential flows. Its core mechanism, **qFold**, a “folding” transform, aggregates many small cryptographic facts, such as keys, evolving application state, policies, delegations, device posture, and protocol transcripts, into a single commitment plus a succinct proof that reveals only what a verifier needs. 

NESSA aims to maintain two compatible backends:
- qfold-ec (Elliptic Curve) is built on Pedersen-style commitments + folded inner-product arguments.
- qfold-pq (Post-Quantum) is built on lattice-based (Module-LWE) commitments and folded lattice sigma-protocols. It is NIST-aligned and provides a migration path for quantum resilience.

The result is an upgradeable privacy layer with consistent APIs, semantics, and proofs across both backends. This allows applications to ship now on elliptic curve cryptography (ECC) while having a clean migration path to post-quantum cryptography (PQC) without re-architecting their privacy model. NESSA emphasizes low overhead, minimal metadata, forward-secure updates, and composability with existing wallets, messengers, and ledgers, all without forcing the adoption of heavy zero-knowledge circuits or exotic infrastructure.

# NESSA

## Our Definition
NESSA is a privacy-first, local-first protocol layer designed for identity, access, and credential management. It enables a user or device to compress numerous sensitive security events, such as enrollments, key rotations, and policy checks, into a single verifiable object. A verifier can inspect this object to confirm policy compliance without accessing the underlying event data or leaking metadata. This capability is powered by **qFold**, a folding transform that aggregates event commitments and generates a succinct proof of correctness. NESSA maintains a consistent application experience across two cryptographic backends. **qFold-EC** is an elliptic-curve track available for immediate deployment, while **qFold-PQ** provides a post-quantum migration path. This dual-engine design ensures that the user experience, APIs, and privacy semantics remain stable even as the underlying cryptography evolves.

## Prove Only What Is Necessary
The protocol operates on a principle of selective disclosure by default. Verifiers learn only the final policy result, for example, that multi-factor authentication is satisfied and a device is valid. They do not learn the count, timestamps, or order of the events unless a policy explicitly requires it. This is the central benefit of folding many cryptographic facts into a single object.

## Offensive Privacy and Metadata Minimization
NESSA is designed to deliberately eliminate correlation surfaces. A single folded object replaces numerous linkable artifacts like tokens or logs. The system does not require a central log of raw events, and proofs are scoped to the verifier and are minimally informative. External attestations are treated as opaque inputs to the fold, preventing them from being used as cross-system tracking beacons.

## Local-First and Self-Sovereign Control
The raw transcript of events remains under the user's or device's control. The only durable artifact a verifier requires is the latest folded object and its corresponding proof. This model aligns with a self-sovereign stance where users hold their own authentication materials, and verifiers check them without collecting sensitive history.

## Key Lifecycle with Forward Security
The protocol supports a robust key lifecycle, including progressions from master to session-specific keys and regular updates. Each folding operation refreshes the cryptographic randomness and invalidates old proof openings. Key rotations and account recovery operations are simply new events folded into the state, providing forward secrecy without revealing the frequency of these changes.

## Integrated Access Control Primitives
Critical access control functions like delegation, revocation, and usage limits are treated as first-class events. Delegations are recorded as events, while revocations are managed through accumulators folded into the state object. Usage limits for credentials can be enforced with privately tracked nullifiers, allowing a verifier to see only the final validity status, not the usage history.

## Context-Bound Proofs without Context Leakage
To prevent replay attacks, proofs can be cryptographically bound to a specific context, such as device posture or a time-window label. The binding is secure and verifiable, yet the context values themselves remain hidden from the verifier.

## Unified Surface for Dual Cryptographic Engines
NESSA provides a consistent API and semantic surface that is independent of the underlying cryptographic engine. **qFold-EC** uses Pedersen-style commitments and inner-product arguments on elliptic curves. **qFold-PQ** uses Module-LWE commitments and folded lattice sigma-protocols for post-quantum resilience. This ensures that migration is a managed process, not a complete product rewrite.

## Simplified Verification
A verifier's task is reduced to a single function call, `VerifyFolded(policy, folded_object, proof)`. This call returns a straightforward valid or invalid result, potentially with reason codes. It removes the need for verifiers to pre-fetch logs or reconstruct historical state, making NESSA practical for diverse environments from backend servers to constrained on-chain verifiers.

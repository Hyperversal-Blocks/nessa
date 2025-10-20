# Non-interactive - Encrypted - Self-Sovereign - Session-Auth - Attestation (NESSA)

## Executive Overview
Project **NESSA** is a privacy-first, local-first cryptographic substrate for managing identity, access, and credential flows. Its core mechanism, **qFold**, a “folding” transform, aggregates many small cryptographic facts, such as keys, evolving application state, policies, delegations, device posture, and protocol transcripts, into a single commitment plus a succinct proof that reveals only what a verifier needs. 

NESSA aims to maintain two compatible backends:
- qfold-ec (Elliptic Curve) is built on Pedersen-style commitments + folded inner-product arguments.
- qfold-pq (Post-Quantum) is built on lattice-based (Module-LWE) commitments and folded lattice sigma-protocols. It is NIST-aligned and provides a migration path for quantum resilience.

The result is an upgradeable privacy layer with consistent APIs, semantics, and proofs across both backends. This allows applications to ship now on elliptic curve cryptography (ECC) while having a clean migration path to post-quantum cryptography (PQC) without re-architecting their privacy model. NESSA emphasizes low overhead, minimal metadata, forward-secure updates, and composability with existing wallets, messengers, and ledgers, all without forcing the adoption of heavy zero-knowledge circuits or exotic infrastructure.

## Problem and the Idea
Modern applications accumulate sensitive cryptographic "breadcrumbs" like key rotations, device enrollments, policy changes, and proof-of-possession events. These facts are spread across logs, tokens, and disparate stores. Verifying them later is often brittle, chatty, and privacy-eroding because verifiers ask for raw logs or linkable attestations, while users leak timing, ordering, and identity metadata. At the same time, teams face a strategic dilemma. They can ship on elliptic curve cryptography (ECC) today and risk a painful post-quantum (PQ) migration later, or bet early on post-quantum cryptography (PQC) and pay performance and tooling costs now.

The qfold mechanism addresses both pressures by replacing these scattered artifacts with a single rolling commitment and a succinct proof of well-formedness. Each time the application state changes, the new fact is "folded" into the accumulator, updating its proof. Verifiers check the one object, learn nothing else, and remain agnostic to whether the underlying cryptography is ECC or PQC.

Think of your security story as a stack of receipts you would rather not show individually. The qfold process is a way to compress the entire stack into one sealed envelope. Anyone can check the seal and be convinced all the right receipts are inside and in good order, but they cannot see the receipts themselves. They also cannot see how many receipts there are, their order, or their timestamps. When you add a new receipt, you do not reopen the whole stack. You just refold the envelope in a way that preserves its integrity.

NESSA is the operating model or policy that governs this system. It defines what kinds of receipts may go in, who is 
allowed to fold, when a verifier is allowed to check the seal, and how revocation or delegation works. With qfold-ec,
the seal is made with elliptic-curve math. With qfold-pq, the seal uses post-quantum (lattice-based) math. Either way, 
the verifier’s experience stays the same. Because folding relies on homomorphic commitments plus a lightweight proof 
system, it is fast enough for mobile and edge devices, simple to audit, offline compatible, and compatible with common 
threat 
models.

## Mathematical Model (Scheme-Agnostic)
We assume an additively homomorphic commitment scheme $C$ that is both binding and computationally hiding.
- **Commit**: The function $C(m;r)$ maps a message $m$ and randomness $r$ to a group element for elliptic curve schemes 
or a lattice vector modulo $q$ for post-quantum schemes.
- **Additivity**: $C(m_1; r_1) \oplus C(m_2; r_2) = C(m_1 + m_2; r_1 + r_2)$
- **Scaling**: For a scalar $\alpha$, the equation $\alpha \cdot C(m; r) = C(\alpha m; \alpha r)$ is defined in the scheme’s algebra.

A transcript consists of a series of events $\{e_i\}_{i=1}^N$. Each event is encoded into the message algebra as $m_i = \mathsf{Enc}(e_i)$. Per-item zero-knowledge relations $\mathcal{R}_i$ enforce that each event is "well-formed under policy $\mathcal{P}$".

Challenges are derived incrementally using the Fiat-Shamir heuristic as $\alpha_i = \mathsf{Chal}(C_1, \dots, C_i, \mathrm{tags})$. The tags include domain separation information such as an application ID, policy hash, epoch, and purpose.

The folded commitment is a single value $C^\star$ calculated as a random linear combination of the individual commitments. 

$$C^\star = \bigoplus_{i=1}^{N} \alpha_i \cdot C(m_i; r_i) = C\left(\sum_i \alpha_i m_i ; \sum_i \alpha_i r_i\right)$$ 

The prover supplies a folded proof $\pi$. This proof demonstrates that the hidden vector $\sum_i \alpha_i m_i$ is a valid linear combination of individually well-formed encodings $m_i \in L_{\mathcal{P}}$ and lies in the language induced by the policy. This is done without revealing any individual $m_i$ or the total number of events $N$. This can be realized with a sigma-protocol folded via inner-product arguments.

A verifier checks the final object and either accepts or rejects it. The check is represented as $(C^{\star}, \pi, \mathsf{pk}, \mathcal{P}) \rightarrow \{\text{accept, reject}\}$.

Soundness reduces to the binding property of the commitment scheme $C$ and the knowledge soundness of the folded proof. Zero-knowledge follows from the simulators for the per-item relation and the folding transform.

## qFold-EC (Elliptic Curve Path)
The qfold-ec instantiation uses Pedersen-style commitments over a prime-order elliptic curve group $\mathbb{G}$. For a scalar encoding $m$ with randomness $r$, the commitment is $C(m;r) = rG + mH$. For vector data, it uses multi-Pedersen commitments of the form $C(\mathbf{m}; \mathbf{r}) = \sum_j (r_j G_j + m_j H_j)$. Generator independence is obtained through methods like hash-to-curve or a verifiable base derivation.

### Folding and Proofs
The folding process uses multi-scalar multiplication (MSM) to compute the final folded commitment $C^\star$ efficiently in a single operation. The prover also generates a proof $\pi$, which is a folded sigma or inner-product argument, similar in style to Bulletproofs. This proof demonstrates that $C^\star$ opens to a hidden vector that satisfies a set of linear constraints. These constraints capture the specific policy $\mathcal{P}$.

Example policies that can be enforced include:
- Device enrollment attested by a registered key.
- Key rotation with monotone versioning.
- Session establishment under 
multi-factor authentication (MFA).
- Selective disclosure of true or false claim predicates without revealing underlying values.

Optional range or bit-decomposition constraints can also be enforced on hidden values. Verification is fast on commodity hardware, requiring only a handful of MSM operations and checks.

### Security and Integration
The security of the scheme reduces to the discrete logarithm assumption on the chosen curve and the knowledge soundness of the inner-product argument. Important engineering practices include using constant-time scalar arithmetic and deterministic nonce derivation for any adjacent cryptographic steps. The choice of curve, such as Ristretto/Ed25519 or **P-256**/secp256r1, depends on the specific deployment requirements.

For a potential integration, qFold-EC could be used as a plug-in verifier in an authentication MVP. A server-side engine would verify folded proofs accompanying register or login requests. An in-memory store would keep only the latest folded object and per-user policy state.

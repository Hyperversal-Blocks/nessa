# Non-interactive - Encrypted - Self-Sovereign - Session-Auth - Attestation (NESSA)

## Executive Overview
Project **NESSA** is a privacy-first, local-first cryptographic substrate for managing identity, access, and credential flows. Its core mechanism, **qFold**, a “folding” transform, aggregates many small cryptographic facts, such as keys, evolving application state, policies, delegations, device posture, and protocol transcripts, into a single commitment plus a succinct proof that reveals only what a verifier needs. 

NESSA aims to maintain two compatible backends:
- qfold-ec (Elliptic Curve) is built on Pedersen-style commitments + folded inner-product arguments.
- qfold-pq (Post-Quantum) is built on lattice-based (Module-LWE) commitments and folded lattice sigma-protocols. It 
  is NIST-aligned and provides a migration path for quantum resilience (out of scope of current document version).

The result is an upgradeable privacy layer with consistent APIs, semantics, and proofs across both backends. This 
allows applications to ship now on elliptic curve cryptography (ECC) while having a clean migration path to 
post-quantum cryptography (PQC) without re-architecting their privacy model. NESSA emphasizes low overhead, minimal metadata, forward-secure updates, and composability with existing wallets, messengers, and ledgers, all without forcing heavy general‑purpose zk circuits. Proofs are specialized to linear policies and can be padded to fixed size.

## Problem and the Idea
Modern applications accumulate sensitive cryptographic "breadcrumbs" like key rotations, device enrollments, policy changes, and proof-of-possession events. These facts are spread across logs, tokens, and disparate stores. Verifying them later is often brittle, chatty, and privacy-eroding because verifiers ask for raw logs or linkable attestations, while users leak timing, ordering, and identity metadata. At the same time, teams face a strategic dilemma. They can ship on elliptic curve cryptography (ECC) today and risk a painful post-quantum (PQ) migration later, or bet early on post-quantum cryptography (PQC) and pay performance and tooling costs now.

The qfold mechanism addresses both pressures by replacing these scattered artifacts with a single rolling commitment and a succinct proof of well-formedness. Each time the application state changes, the new fact is "folded" into the accumulator, updating its proof. Verifiers check the one object, learn nothing else, and remain agnostic to whether the underlying cryptography is ECC or PQC.

Think of your security story as a stack of receipts you would rather not show individually. The qfold process is a 
way to compress the entire stack into one sealed envelope. Anyone can check the seal and be convinced all the right receipts are inside and in good order, but they cannot see the receipts themselves. By default, the verifier does not learn the exact count or order of receipts. Implementations may optionally pad proofs to a fixed size; otherwise, the proof size may grow logarithmically with the number of receipts, which can reveal a coarse upper bound on the count.

NESSA is the operating model or policy that governs this system. It defines what kinds of receipts may go in, who is
allowed to fold, when a verifier is allowed to check the seal, and how revocation or delegation works. With qfold-ec,
the seal is made with elliptic-curve math. With qfold-pq, the seal uses post-quantum (lattice-based) math. Either way,
the verifier’s experience stays the same. Because folding relies on homomorphic commitments plus a lightweight proof
system, it is fast enough for mobile and edge devices, simple to audit, offline compatible, and compatible with common
threat
models.

## Core Principles of the NESSA Architecture

The NESSA architecture is founded on a set of core principles designed to provide robust, future-proof, and practical privacy for modern applications. These pillars guide the system's design and define its value proposition.

### Privacy by Design

NESSA's primary objective is to **prove only what is necessary**. The system operates on a principle of selective disclosure by default. A verifier learns only the outcome of a policy check, such as "MFA is satisfied and the device is valid," without accessing the count, timestamps, or order of the underlying events unless explicitly required by the policy. This is achieved by folding many cryptographic facts into a single, opaque object.

This approach embodies **offensive privacy** through deliberate metadata minimization. The system is designed to eliminate correlation surfaces by replacing numerous linkable artifacts like tokens and logs with a single folded object. No central log of raw events is needed. Proofs are scoped to the verifier and are minimally informative. External attestations are treated as inputs to the fold, preventing them from becoming cross-system tracking beacons.

### Self-Sovereign Control

The architecture is **local-first**, ensuring that the user or device retains control over the raw event transcript. The only durable artifact a verifier requires is the latest folded object and its corresponding proof. This model aligns with a self-sovereign identity stance where users hold the primary materials for authentication, and verifiers check them without collecting or storing the underlying sensitive history.

### Forward-Secure Lifecycle Management

NESSA provides a robust model for the **key lifecycle** with inherent **forward security**. It supports progressions from master keys to medium-term and session-specific keys. Each folding operation refreshes the cryptographic randomness within the accumulator, which invalidates old proof openings. Key rotations and account recovery operations are simply treated as new events folded into the state, providing forward secrecy without revealing the frequency of these changes.

### Integrated Access Control Primitives

The system treats critical access control functions as **first-class events**. Delegations, revocations, and usage limits are not external mechanisms but are folded directly into the user's state object. For instance, credential usage can be limited to a single time or N times through privately tracked nullifiers. The verifier only sees the final validity status of a credential, not its usage history.

### Context-Bound Proofs

To prevent replay attacks, proofs can be **cryptographically bound to a specific context** without leaking that context. This binding can be tied to device posture, a time-window label, or another application-defined tag. The cryptographic binding is secure and verifiable, but the context values themselves remain hidden from the verifier.

### Upgradeable Cryptographic Agility

NESSA provides a consistent API and semantic surface that is independent of the underlying cryptographic engine. The system is maintained with two backends.
* **qFold-EC** uses Pedersen commitments and inner-product arguments on elliptic curves.
* **qFold-PQ** uses Module-LWE commitments and folded lattice sigma-protocols for post-quantum resilience.
  This dual-engine design ensures that migrating from elliptic curve to post-quantum cryptography is a managed process, not a complete product rewrite.

### Simplified Verification

The verifier's task is reduced to a single, simple function call like `VerifyFolded(policy, folded_object, proof)`. This call returns a straightforward valid or invalid result, potentially with reason codes, without requiring the verifier to pre-fetch logs or reconstruct historical state. This simplicity makes NESSA practical for a wide range of environments, from backend servers to mobile applications and even constrained on-chain verifiers.

## Terminology

**NESSA**
> An acronym for **N**on-interactive • **E**ncrypted • **S**elf-**S**overeign • **S**ession-**A**uth • **A**ttestation. It refers to the complete operating model and platform.

**qFold**
> An abbreviation for **quaternion Fold**. The name refers to the core folding transform. The "q" reflects the 
> design intuition drawn from quaternion and spinor algebra, where complex states compose before a single verification check. The name is an analogy inspired by the compositional nature of quaternion algebra and is not related to quantum computing or the use of quaternions in the protocol itself.

## Mathematical Model (Scheme-Agnostic)
We assume an additively homomorphic commitment scheme $C$ that is both binding and computationally hiding.
- **Commit**: The function $C(m;r)$ maps a message $m$ and randomness $r$ to a group element for elliptic curve schemes 
or a lattice vector modulo $q$ for post-quantum schemes.
- **Additivity**: $C(m_1; r_1) \oplus C(m_2; r_2) = C(m_1 + m_2; r_1 + r_2)$
- **Scaling**: The commitment scheme supports public linear combinations, allowing a commitment to be combined with 
  a public scalar $\alpha \in \mathbb{Z}_q$, where $\mathbb{Z}_q$ is the scalar field of the elliptic curve group. 
  - For the **`qfold-ec`** instantiation using Pedersen commitments, this operation is defined as standard scalar multiplication over the elliptic curve. This follows directly from the homomorphic properties of the scheme. The operation is expressed as:
  $$
  \alpha \odot C(m; r) := \alpha \cdot C(m; r) = C(\alpha m; \alpha r)
  $$
  Multiplying the commitment by a scalar is equivalent to creating a new commitment to the scalar-multiplied message using scalar-multiplied randomness.

A transcript consists of a series of events $\{e_i\}_{i=1}^N$. Each event is encoded into the message algebra as $m_i = \mathsf{Enc}(e_i)$. Per-item zero-knowledge relations $\mathcal{R}_i$ enforce that each event is "well-formed under policy $\mathcal{P}$".

### Transcript Accumulation and Challenge Generation

The protocol begins by constructing a transcript accumulator to cryptographically bind the sequence of events. Let 
`tags` be a data structure carrying domain separation information, including the protocol version, application ID, 
policy hash, epoch, and purpose. A collision-resistant hash function $H$ is used to build a sequential hash chain.
The `tags` structure is public and transmitted in the folded object’s header so that verifiers can reconstruct $(R_0)$ and check transcript binding.

The initial state of the accumulator is defined as:
$$
R_0 = H(\text{"NESSA/v1"} \parallel \text{tags})
$$
For each subsequent event commitment $C_i$ in the sequence from $i=1$ to $N$, the accumulator is updated iteratively.
$$
R_i = H(R_{i-1} \parallel C_i \parallel i)
$$
The final accumulator value, or **transcript root**, is $R := R_N$. This single hash value serves as a succinct commitment to the entire ordered sequence of events.

From this transcript root $R$, the public challenges $\alpha_i$ required for the folding process are derived deterministically.
$$
\alpha_i = H_{\text{to\_scalar}}(\text{"NESSA/v1/alpha"} \parallel R \parallel i)
$$
Here, $H_{\text{to\_scalar}}$ is a hash function that outputs a scalar value suitable for operations within the chosen cryptographic group.


### The Folded Commitment

The individual event commitments $\{C_i\}_{i=1}^N$ are aggregated into a single folded commitment $C^{\star}$. This is achieved by computing a random linear combination using the challenges $\alpha_i$.
$$
C^{\star} \gets \bigoplus_{i=1}^{N} \alpha_i \odot C(m_i; r_i)
$$
The symbol $(\odot)$ denotes the scheme’s public linear-combination operator (scalar multiplication in EC; addition plus explicit re-randomization in PQ). Due to the homomorphic properties of the commitment scheme, there exists an aggregate randomness $r^{\star}$ such that the folded commitment correctly opens to the folded message.
$$
C^{\star} = C\left(\sum_i \alpha_i m_i; r^{\star}\right)
$$

### Proof of Correctness

The prover outputs a tuple $(R, C^{\star}, \pi)$, where $\pi$ is a succinct, non-interactive zero-knowledge proof. This proof $\pi$ demonstrates the existence of a set of witnesses $\{(m_i, r_i, C_i)\}_{i=1}^N$ that are committed to by the transcript root $R$ and satisfy two critical properties.

1.  The folded commitment $C^{\star}$ is the correctly computed linear combination of the individual commitments $C_i$, where the challenges $\alpha_i$ are derived from the transcript root $R$ as defined above.
2.  Each individual message encoding $m_i$ is well-formed and valid according to the language of the policy, denoted as $m_i \in L_{\mathcal{P}}$.

This statement is proven using a folded $\Sigma$-protocol. This technique authenticates each commitment $C_i$ against the transcript root $R$ without revealing the underlying messages $m_i$ or the total number of events $N$. A verifier receives `tags`, recomputes $R_0 = H(\text{"NESSA/v1"} \parallel \text{tags})$, verifies that the provided $R$ is consistent with these `tags` via the proof $\pi$, and derives all challenges as $\alpha_i = H_{\text{to\_scalar}}(\text{"NESSA/v1/alpha"} \parallel R \parallel i)$. The verification equation is expressed as:
$$
(R, C^{\star}, \pi, \mathsf{pk}, \mathcal{P}, \text{tags}) \rightarrow \{\text{accept, reject}\}
$$

Soundness reduces to the binding property of the commitment scheme $C$ and the knowledge soundness of the folded proof. Zero-knowledge follows from the simulators for the per-item relation and the folding transform.

## qFold-EC (Elliptic Curve Path)
The `qfold-ec` instantiation uses **Pedersen-style commitments** over a prime-order elliptic curve group $\mathbb{G}$. For a scalar encoding $m$ with randomness $r$, a commitment is formed as $C(m; r) = rG + mH$. This extends naturally to vector data using multi-Pedersen commitments, where $C(\mathbf{m}; \mathbf{r}) = \sum_{j} (r_j G_j + m_j H_j)$.

The generator bases for these commitments are derived using a **"Nothing Up My Sleeve" (a tribute to Satoshi)** 
construction to 
ensure their integrity and independence. Specifically, the base points are generated via a hash-to-curve function with fixed domain separation tags.
$$
G_j = \mathrm{H2C}(\text{"NESSA/v1/base/G"} \parallel \text{curve\_id} \parallel j)
$$
$$
H_j = \mathrm{H2C}(\text{"NESSA/v1/base/H"} \parallel \text{curve\_id} \parallel j)
$$
This verifiable and deterministic process guarantees that no party knows the discrete logarithm relationships between any of the base points. Consequently, under the standard **Discrete Logarithm (DLOG)** assumption in the group $\mathbb{G}$ and the resulting independence of the bases, these commitments are computationally **binding** and **hiding**.
The domain separation strings include `version` and `curve_id`, and the H2C suite identifier is recorded in the public header to make base derivation auditable and unambiguous.

### Folding and Proofs
The folding process uses multi-scalar multiplication (MSM) to compute the final folded commitment $C^\star$ 
efficiently in a single operation. The prover also generates a proof $\pi$, which is a folded sigma or inner-product 
argument, similar in style to Bulletproofs. This proof demonstrates that $C^\star$ opens to a hidden vector that 
satisfies a set of linear constraints. These constraints capture the specific policy $\mathcal{P}$. Conceptually, a 
policy is compiled into a set of linear equations e.g. $(A\mathbf{m}=\mathbf{b})$, that the secret message vectors \
$(\mathbf{m})$ must satisfy; the folded proof then non-interactively attests that the committed messages satisfy 
these equations without revealing $(\mathbf{m})$ or the number of items.


Example policies that can be enforced include:
- Device enrollment attested by a registered key.
- Key rotation with monotone versioning.
- Session establishment under 
multi-factor authentication (MFA).
- Selective disclosure of true or false claim predicates without revealing underlying values.

Optional range or bit-decomposition constraints can also be enforced on hidden values. Verification is fast on commodity hardware, requiring only a handful of MSM operations and checks.

### Security and Integration
The security of the `qfold-ec` instantiation reduces to two foundational assumptions. The first is the hardness of the discrete logarithm problem in the chosen elliptic curve group. The second is the knowledge soundness of the inner-product argument used for the folded proof.

For a secure implementation, it is mandatory to use constant-time scalar arithmetic and deterministic nonce derivation for adjacent cryptographic steps to prevent side-channel attacks. **Group selection:** instantiate over a prime-order group—**Ristretto255** or **secp256r1 (P-256)**. Enforce subgroup checks for secp256r1. Derive commitment base points using a vetted hash-to-curve suite exactly as specified above.
**[WIP]**.

For a potential integration, qFold-EC could be used as a plug-in verifier in an authentication MVP. A server-side engine would verify folded proofs accompanying register or login requests. An in-memory store would keep only the latest folded object and per-user policy state.

## qFold-PQ 
qfold‑pq is out of scope for the objectives outlined for **GG24 Privacy Round** dated **Oct 20th, 2025**. The PQ path will realize folding via addition plus explicit re‑randomization over a Module‑LWE commitment, with challenges derived exactly as in qfold‑ec and with bounded noise growth to preserve completeness and hiding. APIs, headers (`R`, `tags`) and `VerifyFolded` semantics remain identical.

## Threat Model & Properties

- Adversaries: Active network attackers, malicious verifiers, and device compromise with recovery.
- Unforgeability: From commitment binding + proof soundness.
- Zero‑knowledge: Verifier learns only policy results. The proof and header do not reveal the underlying values or the exact count/order/timing of events. If padding is disabled, a logarithmic‑size subcomponent may leak an upper bound on the count; deployments can enable fixed‑size padding to remove this leakage.
- **Non-malleability (ROM)**: The proof binds to $(\text{policy_hash}, \text{tags}, R, \text{version})$ via 
  Fiat-Shamir, where $R$ is the transcript accumulator root. Splicing or mix-and-match attacks across different contexts will fail unless the hash function collides. This property holds in the Random Oracle Model, given the commitment scheme's binding property and the knowledge soundness of the folded argument.
- Forward security: Fresh randomness and epoch tags invalidate old openings; prior states can’t be replayed.
- Side‑channels: Constant‑time scalar arithmetic (EC). Avoid 
data‑dependent memory access in MSM.

## Roadmap and milestones
Near‑term goals are to harden the libraries, publish a short formal note on the fold’s security, and complete an 
audit focused on transcript binding and policy leakage. Milestones include a v0.1 developer preview of qfold‑ec 
suitable for pilot apps. A final milestone for the **privacy round** is a reference verifier that accepts folds via the API and logs no side‑channel metadata.

## Comparison to Alternatives
zkSNARK-based recursion provides powerful generality, but this often comes at the cost of complex circuit authoring, 
high prover memory usage, and trusted setup risks. qfold occupies a leaner niche. It uses linear commitments and 
sigma-protocol folding to achieve most of the benefits of succinctness without a proving system that overfits the 
use cases defined earlier.

Hardware attestation helps with device posture but can leak linkable identity and often requires trusting a vendor. NESSA treats attestations as just another event entering the fold, which deletes linkability at the protocol boundary. While classic batch-verification improves throughput, it does not improve privacy. **qfold improves both**.

## Risks and Mitigations
Folding errors can silently accumulate if challenge derivation is not properly domain-separated or if generators are not set up correctly. We mitigate this risk by embedding domain tags directly into challenges and by deriving generators from a hash-to-curve function or a verifiable seed published in the code.

Side channels are another risk. We mitigate them by using constant-time primitives for all operations involving secret data and by avoiding data-dependent memory access patterns in MSM routines.

## Ethics and Privacy Positioning
NESSA is designed around the principle of data minimization. Proofs reveal exactly what is necessary to authorize a transaction or access, not what happened, when it happened, or on which device. The architecture is local-first, so raw event transcripts do not need to leave the user’s device.

Selective disclosure can be formally scoped, and policies can be audited without requiring access to user data. These properties align directly with the goals of a privacy-focused grants round by providing practical privacy gains for end users and open interfaces for builders.

## Actionable Gaps
- Formal audit and peer-review of the document.
  - An audit pass is required for both the transcript binding and the generator derivation logic to ensure their security and correctness.
- Demo implementation of NESSA and qfold-ec MVP.
  - For performance, the Multi-Scalar Multiplication (MSM) implementation needs to be optimized.
- **The document and the code are still a work in progress.**
- Review of security params.
- A public test harness must be built to rigorously validate the qFold-EC implementation from end to end.

### Common Requirements for Prover and Verifier
The following tasks apply to the overall system.

- The policy encoding, which includes the matrix and constraint description, needs to be finalized. The on-wire 
format must be frozen with versioned domain tags.
- Transcript domain separation must be implemented to distinguish between different contexts, such as application ID, 
policy hash, epoch, and purpose.
- A reference rebase or refresh procedure should be shipped. For the elliptic curve path, this would include an 
optional compaction process if size policies change.

### qFold-EC Specific Tasks
Several key tasks remain for the elliptic curve path.

- The folded Inner Product Argument (IPA) layout needs to be finalized.
- The planned predicate checks for selective disclosure must be implemented.
- A method for vector commitment base derivation needs to be added.
- The data encoders should be subjected to fuzz testing to uncover potential bugs.
- Post paper review, the project must on-board active development team to speed up the development.

### SDK and Demonstration
A draft of the verifier **Software Development Kit (SDK)** should be created for both Go and TypeScript. This **SDK** should expose a consistent `VerifyFolded` function signature across both languages.

Finally, a **demonstration video** should be captured. This video will show a complete fold-to-verify roundtrip to illustrate the system's surface area, even if it uses mock timings.

## Implementation Notes
### Go Authentication MVP Wiring [WIP]
The Go authentication MVP is structured across several packages.

The `pkg/auth/server directory`, containing files like `server.go`, `routes.go`, and `middleware.go`, is responsible for hosting the HTTP services. The `pkg/api` directory translates incoming requests through files such as `register.go`, `login.go`, and `zk.go`. All core qFold plumbing is located in the `pkg/zk` package.

An in-memory data store is provided by `pkg/store/mem.go`, which manages maps for users, the registry, and active sessions. Core data structures like `RegisterContext` and `Session` are defined in `pkg/core/types.go`.

## Appendix A: An Intuitive Example

To build intuition for the folding mechanism, consider a simplified numerical example. Let two distinct events be encoded as small scalar messages. The first event, "device enrolled," is represented by $m_1=3$. The second event, "key rotated to v2," is represented by $m_2=5$. The prover generates commitments to these events as $C_1=C(3;r_1)$ and $C_2=C(5;r_2)$, where $r_1$ and $r_2$ are random values.

The system's transcript absorbs labels and these commitments in sequence. A verifiably random challenge is then derived for each step using the Fiat-Shamir heuristic over the transcript's state. For this example, assume these challenges are $\alpha_1=7$ and $\alpha_2=11$.

The folded commitment $C^{\star}$ is the random linear combination of the individual commitments, defined as $C^{\star} = \alpha_1 C_1 \oplus \alpha_2 C_2$. The calculation proceeds by applying the homomorphic properties of the commitment scheme.

First, the **scaling property**, $\alpha \cdot C(m; r) = C(\alpha m; \alpha r)$, is applied to each term.
$$
7 C_1 = 7 \cdot C(3; r_1) = C(7 \cdot 3; 7r_1) = C(21; 7r_1)
$$
$$
11 C_2 = 11 \cdot C(5; r_2) = C(11 \cdot 5; 11r_2) = C(55; 11r_2)
$$
Next, these scaled commitments are combined using the **additivity property**, $C(m_1; r_1) \oplus C(m_2; r_2) = C(m_1 + m_2; r_1 + r_2)$.
$$
C^{\star} = C(21; 7r_1) \oplus C(55; 11r_2) = C(21 + 55; 7r_1 + 11r_2)
$$
This results in the final folded commitment.
$$
C^{\star} = C(76; r_{\star})
$$
Here, $r_{\star} = 7r_1 + 11r_2$ is the new aggregate randomness. The prover then generates a succinct proof $\pi$. 
This proof demonstrates that the hidden value inside $C^{\star}$, which is 76, is a valid linear combination of 
admissible event encodings. The proof achieves this without revealing the original events, their values, or their 
count. The verifier checks the public header (including `tags` and \(R\)) and the tuple $((R, C^{\star}, \pi))$ 
against the policy $(\mathcal{P}$).

### Appendix B: Security Properties

The security of the system relies on several core cryptographic properties.

#### Binding
The **binding** property of the underlying commitment scheme and the soundness of the folded proof system make the folded object unforgeable. If an adversary could produce two different valid transcripts that decode to the same folded commitment $C^{\star}$, it would constitute a violation of one of these foundational assumptions.

#### Hiding
The system provides **hiding**, a form of zero-knowledge. A verifier interacts with simulators for the per-item relations and the folding transcript, which yield computationally indistinguishable views from a real interaction. Consequently, without the prover's secret witness, the verifier learns only that the folded vector complies with the policy $\mathcal{P}$. The number, content, and order of the individual items remain hidden.

#### Non-malleability
**Non-malleability** is achieved by using the Fiat-Shamir heuristic to derive challenges from the entire prefix of the transcript. This process is strictly domain-separated, binding the sequence of commitments together. This construction ensures that transcripts are non-spliceable and robust against reordering or mix-and-match attacks.

#### Forward Security
The protocol achieves **forward security**. Each fold incorporates fresh randomness into the accumulator, and epoch tagging invalidates stale openings from previous states. Old witness information becomes useless after an update unless its preservation is explicitly allowed by the policy.

### Appendix C: Parameterization and Performance Notes
For the `qFold-EC` instantiation, deployments can use standard, well-vetted elliptic curves. **Ed25519** with the 
**Ristretto255** group is recommended for modern environments due to its performance and security features. For ecosystems requiring specific compliance, **secp256r1** (NIST P-256) is a suitable alternative. Do not instantiate Pedersen over the cofactored edwards25519 group; use Ristretto255 to obtain a prime-order group abstraction, and perform subgroup checks when using secp256r1.

Commitments are single group elements, resulting in a compact size of 32 to 33 bytes. The size of folded proofs depends on the complexity of the policy constraints. For simple linear relations, proofs are typically between 2 KB and 20 KB. The inclusion of additional constraints, such as range proofs, will increase the proof size. All transcripts must include versioned domain tags and a salt bound to the application's unique name and the specific policy hash to ensure cryptographic separation.

### Appendix D: Glossary

**NESSA**
> The operating model, including policies, lifecycle management, and APIs, that operationalizes the `qFold` technology to provide privacy by default.

**qFold**
> A cryptographic folding transform that aggregates many event commitments into a single, compact commitment and an accompanying succinct proof of policy-compliant correctness.

**qFold-EC**
> An instantiation of the `qFold` transform over prime-order elliptic curve groups. It uses Pedersen-style commitments and Bulletproofs-style folded inner-product argument proofs.

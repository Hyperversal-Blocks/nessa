# NESSA qFold-EC: Revised Whitepaper

## Revision Scope

This package defines a corrected, implementation-grade qFold-EC whitepaper. It merges the V3 proof-object structure, $\pi := (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})$ with explicit policy commitments, into the engineering specification while removing non-implementable language suggesting that a verifier can “extract witness coordinates from an IPA opening.” All verifier checks are expressed as explicit proved relations in the discrete-logarithm setting, with deterministic transcript binding and strict wire conformance.

Material changes:

- Transcript accumulation and challenge derivation are now fully verifier-recomputable and byte-exact. Every transcript input is deterministically encoded CBOR under an explicitly frozen restricted profile, with strict reject rules to prevent parser differentials. RFC 8949 requires deterministic encoding choices to be explicit and distinguishes core deterministic map ordering from legacy “length-first” ordering. [1]
- The ciphersuite is frozen to ristretto255 with SHA-512, with point decoding/encoding and scalar guidance pinned to RFC 9496, and hash-to-group / hash-to-field pinned to RFC 9380. [2], [3], [4]
- H2S is defined using RFC 9380 `hash_to_field` with `expand_message_xmd(SHA-512)` and explicit $L = 48$ bytes at 128-bit target security (per RFC 9380’s bias-control formula), with a published DST registry. [2]
- Commitment Profile V2 (two-family, vector blinding) is made dimensionally consistent everywhere. All “scalar $r_\star$” language is removed; folded openings are vectors.
- The proof object is fully specified as $\pi := (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})$. $\pi_{\mathrm{link}}$ is a multi-relation Schnorr NIZK binding the folded commitment $C_\star$ to explicit policy commitments $V_j$. $\pi_{\mathrm{cons}}$ proves folded linear constraints $A \cdot \vec{m}_\star = s \cdot b$ using $\beta$ compression and a Schnorr discrete-log proof on a derived group element $W$. Schnorr NIZK framing and context binding follow RFC 8235 guidance, and deterministic CBOR provides unambiguous boundaries without ad hoc length-prefixing. [1], [5]
- The wire format is defined in CBOR with strict reject behavior (including canonical ristretto decoding and scalar canonicality expectations), and a clear “embedded commitment list” model for v1 to guarantee verifier recomputation.

Open or explicitly unspecified items:

- Any mode that hides the full commitment list $\{C_i\}$ while retaining a verifier-recomputable hash-chain transcript root is blocked unless the accumulator relation is proven inside $\pi$ or replaced by a different accumulator. This remains a major architecture decision.
- Refresh/rebase/compaction semantics remain UNSPECIFIED (design decision: mediated by a blockchain/smart contract, design TBD). No lifecycle security properties that depend on an unspecified state machine are stated as settled.
- Non-linear predicate support inside $\pi$ is not claimed as part of the v1 policy profile (linear-only). A non-linear R1CS folding example is included as a worked test vector and design direction, but a full non-linear proof system mapping (for example Bulletproofs circuits) is left as an explicit open item. Bulletproofs is cited as the canonical discrete-log based circuit-proof reference if that path is chosen. [6]

Implementation status:

- The ECC v1 profile is implementable: deterministic encoding, transcript schedule, base derivation, commitment arithmetic, proof statements, and verifier obligations are specified and cross-checked against primary standards. The included $N = 8$ computed test vectors (linear end-to-end plus a non-linear folding example) validate dimensional correctness and transcript determinism.

## Audit Disposition Ledger

The audit corpus contains the following first-pass and second-pass ledgers and was cross-checked against the archived local PDF authority file and audit update file. [7], [8]

- First-pass audit: “NESSA qFold-EC specification audit and corrected draft (ECC only).pdf” (D-001…D-013).
- Second-pass audit: “Second-Pass Executive Verdict.pdf” (SP-001…SP-010).

Where an audit item is incomplete in the artifact (for example headings without body), it is still dispositioned, but the explanation flags the missing detail and grounds the fix on standards and on consistency with other items.

### Audit ID → disposition table

| Audit ID | Disposition |
| --- | --- |
| D-001 | APPLY AS WRITTEN |
| D-002 | APPLY WITH MODIFICATION |
| D-003 | DEFER / BLOCKED PENDING PROTOCOL DECISION |
| D-004 | APPLY WITH MODIFICATION |
| D-005 | APPLY AS WRITTEN |
| D-006 | APPLY AS WRITTEN |
| D-007 | APPLY AS WRITTEN |
| D-008 | APPLY AS WRITTEN |
| D-009 | APPLY WITH MODIFICATION |
| D-010 | APPLY WITH MODIFICATION |
| D-011 | APPLY AS WRITTEN |
| D-012 | APPLY AS WRITTEN |
| D-013 | DEFER / BLOCKED PENDING PROTOCOL DECISION |
| SP-001 | APPLY AS WRITTEN |
| SP-002 | APPLY AS WRITTEN |
| SP-003 | APPLY AS WRITTEN |
| SP-004 | APPLY WITH MODIFICATION |
| SP-005 | APPLY AS WRITTEN |
| SP-006 | APPLY AS WRITTEN |
| SP-007 | APPLY WITH MODIFICATION |
| SP-008 | APPLY AS WRITTEN |
| SP-009 | DEFER / BLOCKED PENDING PROTOCOL DECISION |
| SP-010 | APPLY AS WRITTEN |

### Per-item ledger entries

#### D-001 — Transcript accumulation and binding is under-specified and ambiguous

Affected sections: transcript root definition; transcript encoding; verifier recomputation.
Audit source(s): first-pass D-001.
Disposition: APPLY AS WRITTEN.
Explanation: The corrected spec defines transcript steps as deterministic CBOR-encoded arrays and hashes the resulting bytes. This eliminates non-injective concatenation ambiguity and provides deterministic, verifier-recomputable challenges. RFC 8949 requires deterministic encoding choices to be explicit and defines deterministic encoding and map ordering options. [1]
Authority basis: RFC 8949 deterministic encoding; FIPS 180-4 SHA-512. [1], [4]
Resulting whitepaper change: Added “Deterministic CBOR profile” and rewrote the transcript schedule with explicit CBOR shapes and strict reject rules.

#### D-002 — Transcript root semantics when commitments are hidden

Affected sections: transcript root semantics; commitment list availability; verifier interface.
Audit source(s): first-pass D-002.
Disposition: APPLY WITH MODIFICATION.
Explanation: v1 requires the full ordered list $C_1,\ldots,C_N$ be embedded on wire. Any “hidden $\{C_i\}$” mode requires a different proof relation (prove the hash-chain relation inside $\pi$ or redesign the accumulator). This is left as an open item rather than a half-spec.
Authority basis: Fiat–Shamir requires verifier-recomputable transcripts; RFC 8235 emphasizes explicit context binding and clear boundaries for hashed items. [5]
Resulting whitepaper change: Normatively fixed “v1 commitment list semantics: embedded list” and moved hidden-$\{C_i\}$ to Open Items.

#### D-003 — Challenge derivation requires a public index domain

Affected sections: $\alpha_i$ schedule; count/order hiding claims.
Audit source(s): first-pass D-003.
Disposition: DEFER / BLOCKED PENDING PROTOCOL DECISION.
Explanation: $\alpha_i$ derivation requires a defined index domain $i \in \{1,\ldots,N\}$ (and thus N is known unless padding is specified). The revised whitepaper removes any claim that count/order are hidden “by default” in v1.
Authority basis: RFC 8235 requires the transcript-defined challenge to be well specified; deterministic CBOR ensures agreement on what is hashed. [1], [5]
Resulting whitepaper change: Rewrote privacy and leakage model; added optional padded-length profile as future work. Count and order hiding are not claimed in v1 unless a padded index domain is specified.

#### D-004 — Domain separation and transcript labels are inconsistent

Affected sections: DST registry; suite binding; transcript labels.
Audit source(s): first-pass D-004.
Disposition: APPLY WITH MODIFICATION.
Explanation: The revised spec publishes both (a) protocol transcript labels and (b) RFC 9380 DSTs for `hash_to_field` and `hash_to_ristretto255`, with explicit registry entries and suite binding. RFC 9380 specifies DST construction and REQUIRED identifier guidance for ristretto255 + XMD:SHA-512. [2]
Resulting whitepaper change: Added a normatively frozen DST registry and required binding of the REQUIRED identifier into tags.

#### D-005 — Commitment model mismatch and dimensional inconsistency

Affected sections: commitment definition; folding equations; notation for $r_\star$.
Audit source(s): first-pass D-005.
Disposition: APPLY AS WRITTEN (design decision fixes to V2).
Explanation: Commitment Profile V2 is used consistently: $\mathrm{Com}(\vec{m}; \vec{r})$ uses vector blinding and two base families. Folded openings are vectors, not scalars.
Authority basis: Linear commitment homomorphism is standard for Pedersen-style commitments; Bulletproofs discusses Pedersen commitments as inputs and relies on discrete log security. [6]
Resulting whitepaper change: Removed scalar-only commitment definitions from normative text; all equations use $\vec{r} \in \mathbb{F}_r^d$.

#### D-006 — Generator/base derivation under-specified

Affected sections: base derivation; suite identifiers; curve identifiers.
Audit source(s): first-pass D-006.
Disposition: APPLY AS WRITTEN.
Explanation: All bases are derived via RFC 9380 `hash_to_ristretto255` (`expand_message_xmd(SHA-512)`, 64 bytes) mapped via `ristretto255_map`, where the map is the ristretto element-derivation function described in RFC 9496. Deterministic retry rules handle identity/duplicates. [2], [3]
Resulting whitepaper change: Added a base-derivation algorithm with explicit messages, DSTs, identity/duplicate checks, and retry counter.

#### D-007 — Point decoding/validation and canonicality are missing

Affected sections: decoding requirements; strict reject behavior.
Audit source(s): first-pass D-007.
Disposition: APPLY AS WRITTEN.
Explanation: The revised spec mandates RFC 9496 decoding for ristretto elements, including the rule that non-canonical values are rejected and the most significant bit is not masked during decode. [3]
Resulting whitepaper change: Added canonical decode requirements, re-encode checks, and reject rules.

#### D-008 — H-to-scalar is informal

Affected sections: H2S definition; $\alpha_i$ schedule; $\beta$ sampling.
Audit source(s): first-pass D-008.
Disposition: APPLY AS WRITTEN.
Explanation: H2S uses RFC 9380 `hash_to_field` over $\mathrm{GF}(r)$ with `expand_message_xmd(SHA-512)`. L is fixed by the RFC formula, and for a 255-bit prime at $k = 128$, $L = 48$. [2]
Resulting whitepaper change: Added normative H2S algorithm, parameters, and DST registry.

#### D-009 — $\pi$ is unspecified / relies on infeasible verification language

Affected sections: proof statement; verifier obligations; policy verification.
Audit source(s): first-pass D-009; second-pass SP-008.
Disposition: APPLY WITH MODIFICATION.
Explanation: The corrected design is $\pi := (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})$ using Schnorr NIZKs (Fiat–Shamir) rather than any verifier “extraction.” RFC 8235 specifies Schnorr NIZK structure, challenge binding to optional “OtherInfo,” and recommends explicit boundaries between concatenated items; deterministic CBOR supplies unambiguous boundaries. [1], [5]
Resulting whitepaper change: Removed all IPA-extraction language; defined NP statements and explicit verifier algorithms.

#### D-010 — Policy compilation and predicate scope are unclear

Affected sections: policy language; compilation; selective disclosure outputs.
Audit source(s): first-pass D-010.
Disposition: APPLY WITH MODIFICATION.
Explanation: v1 restricts policies to linear equalities over $\mathbb{F}_r$ and boolean outputs only. The compiled representation $(A,b)$ is fixed as deterministic CBOR and hash-bound. Non-linear predicates are moved to Open Items with explicit requirements for the proof system if extended. Bulletproofs is cited as the primary discrete-log circuit proof reference if that route is chosen. [1], [6]
Resulting whitepaper change: Added grammar/semantics for the v1 linear profile and a canonical compiled-policy encoding.

#### D-011 — Wire format and reject behavior incomplete

Affected sections: on-wire schema; reject rules; size limits.
Audit source(s): first-pass D-011; second-pass SP-006.
Disposition: APPLY AS WRITTEN.
Explanation: The revised whitepaper provides a strict CBOR schema (integer-keyed maps), deterministic encoding profile, strict rejects, and normative limits. RFC 8949 requires explicit deterministic format choices and documents deterministic map ordering options. [1]
Resulting whitepaper change: Added wire schema tables, strict reject rules, and conformance testing expectations.

#### D-012 — Performance and proof-size claims unsupported

Affected sections: performance section; proof sizing.
Audit source(s): first-pass D-012 (sparse in artifact).
Disposition: APPLY AS WRITTEN.
Explanation: Any numeric performance claims are removed unless tied to explicit parameters $(N,d,k)$ and measured benchmarks. Bulletproofs provides general asymptotic guidance for circuit proof size, but concrete numbers depend on circuit size and implementation choices. [6]
Resulting whitepaper change: Replaced overclaims with parameter dependence and required benchmark/test-vector methodology.

#### D-013 — Refresh/rebase/compaction semantics missing

Affected sections: lifecycle semantics.
Audit source(s): first-pass D-013.
Disposition: DEFER / BLOCKED PENDING PROTOCOL DECISION.
Explanation: Refresh/rebase/compaction semantics are delegated to a protocol blockchain/contract, but contract semantics are TBD. No forward-security or revocation claims depending on this are made as settled in v1.
Authority basis: absent a defined state machine, no primary source can justify claimed properties.

#### SP-001 — Fix commitment formula and $r_\star$ semantics

Affected sections: commitment profile; folding equation.
Audit source(s): second-pass SP-001.
Disposition: APPLY AS WRITTEN.
Explanation: The revised spec uses the vector commitment formula consistently and avoids scalar $r_\star$. [6], [9]
Resulting whitepaper change: Harmonized all commitment arithmetic to V2.

#### SP-002 — Separate transcript labels from RFC 9380 DSTs

Affected sections: transcript labels; DSTs; suite identifiers.
Audit source(s): second-pass SP-002.
Disposition: APPLY AS WRITTEN.
Explanation: The revised spec defines NESSA transcript labels for CBOR-framed transcript messages and RFC 9380 DSTs for `hash_to_field` and `hash_to_ristretto255` separately, preventing accidental collisions and clarifying what is used where. [2]

#### SP-003 — Bind REQUIRED identifier for hash-to-ristretto255

Affected sections: tags schema; suite binding.
Audit source(s): second-pass SP-003.
Disposition: APPLY AS WRITTEN.
Explanation: Tags MUST include `ristretto255_XMD:SHA-512_R255MAP_RO_` exactly, as specified by RFC 9380 for this instantiation. [2]
Resulting whitepaper change: Added required tag field `suite_id_h2g` with that value.

#### SP-004 — P-256 subgroup checks are context-specific

Affected sections: multi-suite roadmap.
Audit source(s): second-pass SP-004.
Disposition: APPLY WITH MODIFICATION.
Explanation: v1 is ristretto-only; P-256 validation rules are referenced only as future extension guidance (SEC 1) and not used in the v1 proof relation.
Authority basis: RFC 9496 defines ristretto as prime-order; SEC 1 applies to Weierstrass curve point encodings for future extension, with SEC 2 and NIST SP 800-186 retained as future domain-parameter references. [3], [10], [11], [12]

#### SP-005 — Use hash-to-field for H2S

Affected sections: H2S definition; $\beta$ compression.
Audit source(s): second-pass SP-005.
Disposition: APPLY AS WRITTEN.
Explanation: `hash_to_field` is used for H2S, with explicit L and k. [2]

#### SP-006 — Bind d, N, and tags into transcript

Affected sections: tags; transcript schedule.
Audit source(s): second-pass SP-006.
Disposition: APPLY AS WRITTEN.
Explanation: d and `policy_hash` live in tags which are hash-bound into R0; N is determined by the embedded list length and thus transcript-bound. [1], [5]

#### SP-007 — Hidden N/order claims are blocked unless padding profile exists

Affected sections: privacy model.
Audit source(s): second-pass SP-007.
Disposition: APPLY WITH MODIFICATION.
Explanation: v1 declares N and order as visible unless padding profile is enabled (future work). [1], [5]

#### SP-008 — $\pi$ must be fully mapped to a checkable verifier schedule

Affected sections: proof system; verifier obligations.
Audit source(s): second-pass SP-008.
Disposition: APPLY AS WRITTEN.
Explanation: $\pi_{\mathrm{link}}$ and $\pi_{\mathrm{cons}}$ are defined with explicit transcript-bound challenges and verifier equations. [5]

#### SP-009 — Hidden $\{C_i\}$ requires new proof relation

Affected sections: hidden commitments mode.
Audit source(s): second-pass SP-009.
Disposition: DEFER / BLOCKED PENDING PROTOCOL DECISION.
Explanation: Not supported in v1; listed as open item requiring either a proof of accumulator relation or redesigned transcript binding.

#### SP-010 — Commitment equation normalization

Affected sections: commitment definition.
Audit source(s): second-pass SP-010.
Disposition: APPLY AS WRITTEN.
Explanation: Commitment equation is fixed to $\sum_{j=1}^{d}(r_j B^r_j + m_j B^m_j)$ with explicit base derivation and vector blinding. [6], [9]

## NESSA qFold-EC Protocol Specification

### Executive Summary

qFold-EC is a transcript-bound folding construction over prime-order elliptic-curve commitments. A prover commits to a sequence of fixed-width event encodings $m_i \in \mathbb{F}_r^d$ using a two-family, vector-blinded Pedersen-style commitment, derives Fiat–Shamir challenges $\alpha_i$ from a canonical transcript root $R$, folds event commitments into a single folded commitment $C_\star$, and proves that the folded witness satisfies a declared linear policy without revealing witness values.

This v1 specification is implementation-grade for the linear-only policy profile. It freezes:

- Group: ristretto255 with canonical decoding/encoding and scalar guidance per RFC 9496. [3]
- Hash: SHA-512 per FIPS 180-4. [4]
- Hash-to-group and hash-to-field: RFC 9380 `hash_to_ristretto255` and `hash_to_field` with `expand_message_xmd(SHA-512)`, including bias-control length L. [2]
- Canonical encoding and transcript binding: deterministic CBOR per RFC 8949 with an explicit restricted profile and strict reject rules. [1]
- Proof: $\pi := (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})$, with explicit policy commitments $V_j$, multi-relation Schnorr linkage, and compressed linear constraint proof. Schnorr NIZK framing and contextual binding follow RFC 8235. [3], [5]

Any older text implying verifiers can “extract coordinates from an IPA opening” is intentionally absent. If a future version adopts a Bulletproofs arithmetic-circuit proof, the mapping must be explicit and complete (statement, witness, transcript schedule, and binding points). Bulletproofs is the primary discrete-log circuit-proof reference for that path. [6]

### Scope and Threat Model

The adversary controls prover inputs and wire encodings and attempts to:

- cause transcript divergence across implementations (parser differentials, non-deterministic encoding),
- exploit non-canonical point/scalar encodings,
- substitute policy, suite, or encoding identifiers after proof generation,
- force verification DoS via oversized or pathological inputs.

Security assumptions:

- Discrete logarithm hardness in ristretto255 and random-oracle Fiat–Shamir for Schnorr NIZKs. [3], [5]
- Hash-to-field and hash-to-group behave as specified in RFC 9380, including bias control via L. [2]

Declared leakage in v1:

- The verifier learns N and the order of commitments (because the commitment list is embedded and indexed for $\alpha_i$ derivation). Count/order hiding requires a padded profile (Open Items). [1], [5]

### Notation

- $\mathbb{G}$: ristretto255 group (prime order $r$). [3]
- $\mathbb{F}_r$: scalar field $\mathrm{GF}(r)$.
- Scalars are integers modulo r, encoded as 32-byte little-endian. RFC 9496 recommends rejecting non-canonical scalar encodings when parsing scalars and notes omitting strict checks is not recommended. [3]
- Points are 32-byte ristretto encodings; decoding rejects non-canonical values and does not mask the MSB. [3]

### Ciphersuite and Suite Binding

This specification fixes:

- $H_{\mathrm{tr}} = \mathrm{SHA512}$, denoting SHA-512 from FIPS 180-4. [4]
- H2G is RFC 9380 `hash_to_ristretto255` (`expand_message_xmd(SHA-512)`, output length 64, then `ristretto255_map`). RFC 9380 defines this construction and states the REQUIRED identifier for this instantiation. [2]
- REQUIRED identifier bound in tags: `ristretto255_XMD:SHA-512_R255MAP_RO_`. [2]

H2S uses RFC 9380 `hash_to_field` over $\mathrm{GF}(r)$:

- Bias control: RFC 9380 defines $L = \lceil(\lceil\log_2(p)\rceil + k)/8\rceil$ and gives $L = 48$ bytes for a 255-bit prime at $k = 128$. For ristretto scalars, set $k = 128$ and $L = 48$. [2]

### Deterministic CBOR Profile and Strict Reject Rules

All protocol-layer hashed inputs (tags, transcript messages, proof transcript inputs) MUST be deterministically encoded CBOR under a restricted profile.

Deterministic encoding choice:

- Use RFC 8949 “core deterministic encoding requirements” (map keys sorted by bytewise lexicographic order of deterministic encodings). RFC 8949 notes legacy “length-first” ordering exists for compatibility, but this spec does not use it in v1. [1]
- If JSON is used as a non-normative authoring layer before CBOR, RFC 8785 is retained only as optional JSON canonicalization guidance. [13]

Restricted profile (v1):

- Allowed CBOR types: unsigned/negative integers, byte strings, text strings, arrays, maps, simple values true/false/null.
- Forbidden: floats, NaN/Infinity, tags (major type 6), indefinite-length items, duplicate map keys.
- Protocol maps MUST use unsigned integer keys only (major type 0).
- Cryptographic encodings:
  - ristretto point: bstr length 32,
  - scalar: bstr length 32,
  - SHA-512 digest: bstr length 64.

Strict reject behavior:

- Verifiers MUST reject inputs not deterministically encoded under this profile. RFC 8949 stresses protocols must be explicit about deterministic encoding choices. [1]
- Verifiers MUST reject ristretto point encodings that fail RFC 9496 decode requirements. [3]
- Verifiers SHOULD reject non-canonical scalars (RFC 9496 guidance). [3]
- Verifiers MUST reject unknown top-level keys (closed-world schema for v1), missing required keys, and oversized arrays beyond configured limits (DoS control).

### Tags Schema

Tags τ is a deterministic CBOR map with unsigned integer keys.

| Key | Name | Type | Requirement |
| ---: | --- | --- | --- |
| 0 | `protocol_version` | uint | MUST be 1 |
| 1 | `protocol_suite_id` | tstr | MUST be `NESSA-EC-RISTRETTO255-SHA512-v1` |
| 2 | `h2g_required_id` | tstr | MUST be `ristretto255_XMD:SHA-512_R255MAP_RO_` [2] |
| 3 | `encoding_id` | tstr or uint | application-defined |
| 4 | `encoding_hash` | bstr(64) | SHA-512 of canonical encoding schema |
| 5 | d | uint | message dimension |
| 6 | `policy_id` | tstr or uint | application-defined |
| 7 | `policy_hash` | bstr(64) | SHA-512 of compiled policy object |
| 8 | `k_rows` | uint | number of linear constraints |
| 9 | `transcript_seed` | bstr(64) | OPTIONAL; hash of signature/proof-chain digest |

`tags_hash := SHA-512(EncCBOR(τ))`. In formulas below, `SHA512(...)` denotes SHA-512 applied to bytes. [1], [4]

Implementation note: this repository currently serializes `encoding_id` and `policy_id` as application-defined strings for interoperability with higher-level APIs, while preserving the same keyed map structure and hashing rules.

### Commitment Profile V2

Fix d per `encoding_id`.

Base families:

- $\{B^r_j\}_{j=1}^{d}$ blinding bases,
- $\{B^m_j\}_{j=1}^{d}$ message bases,
- policy bases $G_{\mathrm{pol}}$ and $H_{\mathrm{pol}}$.

Base derivation (normative):

For each base label and index j, derive a ristretto point via:

$$
P := \mathrm{hash\_to\_ristretto255}(\mathrm{msg}, \mathrm{DST})
$$

where `hash_to_ristretto255` is RFC 9380 (`expand_message_xmd(SHA-512)`, 64 bytes, then `ristretto255_map`). [2], [3]

Derivation messages use deterministic CBOR:

- `msg(B^r_j) = EncCBOR(["base","Br",j,protocol_suite_id])`
- `msg(B^m_j) = EncCBOR(["base","Bm",j,protocol_suite_id])`
- `msg(G_pol) = EncCBOR(["base","Gpol",0,protocol_suite_id])`
- `msg(H_pol) = EncCBOR(["base","Hpol",0,protocol_suite_id])`

DSTs are fixed ASCII strings (see DST registry below).

Deterministic retry:

- If the derived encoding equals the identity encoding (32 zero bytes; see RFC 9496 generator-multiple test vectors including `B[0]`) or duplicates an already-derived base in the same family, append a `u32` counter to `msg` and retry until valid. [2], [3]

Commitment:

For $\vec{m} \in \mathbb{F}_r^d$ and $\vec{r} \in \mathbb{F}_r^d$,

$$
\mathrm{Com}(\vec{m}; \vec{r}) :=
\sum_{j=1}^{d} (r_j B^r_j + m_j B^m_j) \in \mathbb{G}.
$$

### Event Encoding `Enc(e)`

$\mathrm{Enc}(e)$ outputs $\vec{m} \in \mathbb{F}_r^d$ in fixed width.

Schema requirements per `encoding_id`:

- list of event types,
- ordered fields and domains,
- per-field mapping to scalars with reject rules,
- interpretation (human semantics),
- dimension d,
- canonical schema encoding for hashing to `encoding_hash`.

Example (illustrative):

$d = 4$ fields: $m_0,m_1,m_2,m_3$ where $m_0 + m_1$ encodes a total and $(m_2,m_3)$ encode a relation. Out-of-domain values MUST be rejected before reduction modulo $r$ to prevent wrap-around semantics.

### Transcript Schedule

Commitment list semantics (v1):

- The verifier MUST be given the full ordered list $C_1,\ldots,C_N$ embedded on wire as 32-byte ristretto encodings. Hidden $\{C_i\}$ modes are not v1. [1], [5]

Transcript root:

Let `tags_hash` be as above. Let $\mathrm{Enc}(C_i)$ be the 32-byte encoding of $C_i$.

- $R_0 := \mathrm{SHA512}(M_{R0})$, where `M_R0 = EncCBOR(["NESSA-EC:v1:R0", tags_hash])`.
- For $i = 1,\ldots,N$: $R_i := \mathrm{SHA512}(M_{Ri})$, where `M_Ri = EncCBOR(["NESSA-EC:v1:Ri", i, R_{i-1}, Enc(C_i)])`.
- $R := R_N$

Challenges:

- $\alpha_i := \mathrm{H2S}(\mathrm{DST\_ALPHA}, M_{\alpha,i}) \in \mathbb{F}_r$, where `M_alpha_i = EncCBOR(["alpha", R, i])`.

H2S is RFC 9380 `hash_to_field` with $L = 48$ bytes and `expand_message_xmd(SHA-512)`. [2]

Folding:

- $C_\star := \sum_{i=1}^{N} \alpha_i C_i$
- $\vec{m}_\star := \sum_{i=1}^{N} \alpha_i \vec{m}_i$
- $\vec{r}_\star := \sum_{i=1}^{N} \alpha_i \vec{r}_i$

### Policy Language v1

v1 policies are linear equalities:

$$
A \cdot \vec{m} = b,
\quad A \in \mathbb{F}_r^{k \times d},
\quad b \in \mathbb{F}_r^k.
$$

Per-event policy:

For each $i$, $A \cdot \vec{m}_i = b$.

Folded policy:

Let $s := \sum_{i=1}^{N} \alpha_i$. Then $A \cdot \vec{m}_\star = s \cdot b$ if all per-event constraints hold.

Compiled policy encoding (normative):

A compiled policy object is a deterministic CBOR map:

- 0: `version` (`uint` value `1`)
- 1: $d$ (`uint`)
- 2: `k_rows` (uint)
- 3: $A$ (array of `k_rows` arrays of $d$ scalars, each scalar as `bstr32` LE)
- 4: $b$ (array of `k_rows` scalars, each scalar as `bstr32` LE)

`policy_hash := SHA-512(EncCBOR(policy_compiled))`. [1], [4]

### Proof System $\pi := (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})$

All Schnorr NIZKs are Fiat–Shamir transformed with challenges derived from deterministic CBOR encodings that bind `tags_hash` and the full public statement context. RFC 8235 requires context (“OtherInfo”) formatting be fixed and explicitly defined and recommends clear boundaries between concatenated items; CBOR provides unambiguous boundaries. [1], [5]

Transcript-engineering libraries such as Merlin may inform implementations, but they do not replace the explicit CBOR transcript schedule defined here. [14]

#### Explicit Policy Commitments

For $j = 1,\ldots,d$, define:

$$
V_j := \gamma_j G_{\mathrm{pol}} + m_{\star,j} H_{\mathrm{pol}},
$$

where $\gamma_j \in \mathbb{F}_r$ is fresh blinding.

#### $\pi_{\mathrm{link}}$: Linkage Proof

Public statement:

$(\mathrm{tags\_hash}, R, C_\star, V_1,\ldots,V_d, B^r_j, B^m_j, G_{\mathrm{pol}}, H_{\mathrm{pol}})$.

Witness:

$(\vec{m}_\star, \vec{r}_\star, \vec{\gamma})$ such that:

- $C_\star = \mathrm{Com}(\vec{m}_\star; \vec{r}_\star)$
- For all $j = 1,\ldots,d$: $V_j = \gamma_j G_{\mathrm{pol}} + m_{\star,j} H_{\mathrm{pol}}$

Protocol (Schnorr multi-relation):

Prover commits to random `a_m`, `a_r`, and `a_gamma`, computes `T_C` and `T_Vj`, derives `c_link` via H2S, returns z values. Verifier checks the standard Schnorr equations per relation.

Challenge:

$$
c_{\mathrm{link}} :=
\mathrm{H2S}(\mathrm{DST\_LINK}, M_{\mathrm{link}}).
$$

Here `M_link = EncCBOR(["link", tags_hash, R, Enc(C_star), Enc(V_1), ..., Enc(V_d), Enc(T_C), Enc(T_V_1), ..., Enc(T_V_d)])`.

#### $\pi_{\mathrm{cons}}$: Linear Constraints Proof via Compression

Goal: prove $A \cdot \vec{m}_\star = s \cdot b$ without revealing $\vec{m}_\star$.

Verifier derives $\beta_l$ for $l = 1,\ldots,k$:

$$
\beta_l := \mathrm{H2S}(\mathrm{DST\_BETA}, M_{\beta,l}).
$$

Here `M_beta_l = EncCBOR(["beta", R, policy_hash, l])`.

Compute:

- $v := \beta^{\mathsf T} A \in \mathbb{F}_r^d$
- $u := \beta^{\mathsf T}(s \cdot b) \in \mathbb{F}_r$

Define:

$$
W := \sum_{j=1}^{d} v_j V_j - u H_{\mathrm{pol}}.
$$

If $\sum_{j=1}^{d} v_j m_{\star,j} = u$, then $W = (\sum_{j=1}^{d} v_j \gamma_j)G_{\mathrm{pol}}$, so $W$ lies in the span of $G_{\mathrm{pol}}$.

$\pi_{\mathrm{cons}}$ is a Schnorr proof of knowledge of $w$ such that:

$$
W = w G_{\mathrm{pol}}.
$$

Challenge:

$$
c_{\mathrm{cons}} :=
\mathrm{H2S}(\mathrm{DST\_CONS}, M_{\mathrm{cons}}).
$$

Here `M_cons = EncCBOR(["cons", tags_hash, R, policy_hash, Enc(W), Enc(T)])`.

Verifier check:

$$
z G_{\mathrm{pol}} \stackrel{?}{=} T + c_{\mathrm{cons}} W.
$$

RFC 8235 defines Schnorr NIZK structure and requirements for OtherInfo formatting and boundaries. [5]

### Wire Format

Top-level object is deterministic CBOR map with unsigned integer keys:

| Key | Field | Type |
| ---: | --- | --- |
| 0 | `tags` | map |
| 1 | `commitments` | array of bstr32 (`C_i` encodings) |
| 2 | `transcript_root` | bstr64 (`R`) |
| 3 | `folded_commitment` | bstr32 (`Enc(C_star)`) |
| 4 | `policy_commitments` | array length $d$ of bstr32 (`Enc(V_j)`) |
| 5 | `proof` | map |

Proof object:

- `proof[0] = pi_link`, `proof[1] = pi_cons`.

The exact internal CBOR maps for $\pi_{\mathrm{link}}$ and $\pi_{\mathrm{cons}}$ are specified in Appendix A (conformance schema).

### Verifier Algorithm (Normative)

Given wire object:

1. Parse CBOR and enforce restricted deterministic profile; reject otherwise. [1]
2. Decode each $C_i$, $C_\star$, $V_j$, and proof point using RFC 9496 decode; reject failures and non-canonical values. [3]
3. Recompute `tags_hash`, $R_0,\ldots,R_N$, and $R$; reject if mismatch.
4. Derive $\alpha_i$ and recompute $C_\star$; reject if mismatch.
5. Verify $\pi_{\mathrm{link}}$.
6. Compute $\beta$, $v$, $u$, and $W$ and verify $\pi_{\mathrm{cons}}$.
7. Accept iff all checks pass.

### Security Considerations

- Deterministic encoding and strict rejects are part of the security boundary. RFC 8949 requires protocols to be explicit about deterministic encoding choices; failing to do so risks transcript divergence. [1]
- Ristretto decoding must reject non-canonical values and does not mask the MSB (decode differs from RFC 7748 field decoding). [3]
- RFC 9380 `hash_to_field` controls modulo bias via $L$; this spec fixes $k = 128$ and $L = 48$ for $\mathrm{GF}(r)$. [2]
- Schnorr NIZKs require fixed context binding; RFC 8235 emphasizes that OtherInfo formatting must be fixed and that item boundaries must be clear. Deterministic CBOR addresses boundary clarity while minimizing ad hoc length-prefixing. [1], [5]
- DoS bounds: verifiers SHOULD enforce maximum $N$, maximum $d$, and maximum CBOR size; these are deployment profile parameters and must be documented per application.

### Appendix A: DST Registry and CBOR Conformance Schema

DST registry (ASCII):

- `DST_ALPHA` = `NESSA-EC:v1:alpha`
- `DST_BETA` = `NESSA-EC:v1:beta`
- `DST_LINK` = `NESSA-EC:v1:link`
- `DST_CONS` = `NESSA-EC:v1:cons`
- base DSTs: `NESSA-EC:v1:base:Br`, `...:Bm`, `...:Gpol`, `...:Hpol`

CBOR schemas (v1):

- `tags`: map keys 0..9 as specified.
- $\pi_{\mathrm{link}}$: map `{0:T_C, 1:T_V_array, 2:z_r_array, 3:z_m_array, 4:z_gamma_array}`.
- $\pi_{\mathrm{cons}}$: map `{0:T, 1:z}`.

Arrays must have exact lengths ($d$, `k_rows`) and scalars must be `bstr32`.

### Appendix B: Deterministic CBOR Examples and Sample Transcripts

RFC 8949 deterministic encoding uses core deterministic ordering by lexicographic ordering of deterministic key encodings; v1 fixes that and forbids length-first ordering. [1]

Example tags (CBOR diagnostic, from Test Vector TV-LIN-8 below):

```cbor-diag
{
  0: 1,
  1: "NESSA-EC-RISTRETTO255-SHA512-v1",
  2: "ristretto255_XMD:SHA-512_R255MAP_RO_",
  3: 1,
  4: h'…64 bytes…',
  5: 4,
  6: 1,
  7: h'…64 bytes…',
  8: 2
}
```

$R_0$ preimage (CBOR diagnostic):

```cbor-diag
["NESSA-EC:v1:R0", tags_hash]
```

$R_1$ preimage:

```cbor-diag
["NESSA-EC:v1:Ri", 1, R0, C1]
```

### Appendix C: Computed Test Vectors ($N = 8$)

The following vectors are generated directly from `impl/test_vectors.py` and serialized in `impl/test_vectors_output.json` using the corrected transcript schedule and challenge labels.

#### TV-LIN-8: Linear Policy End-to-End ($\pi_{\mathrm{link}} + \pi_{\mathrm{cons}}$)

Parameters:

- $d = 4$
- `k_rows` = 2
- $N = 8$
- Policy rows = `[[1, 1, 0, 0], [0, 0, 3, -1]]`
- Policy targets = `[1000, 7]`

Public wire inputs (hex):

- `tags_cbor`: `aa000101781f4e455353412d45432d52495354524554544f3235352d5348413531322d763102782472697374726574746f3235355f584d443a5348412d3531325f523235354d41505f524f5f036854562d4c494e2d38045840287034d3d090faf7a57d7c4a812225e8026fb12e6c08a0f1e07dc4dc0977d2f43da658add3b4ecff118f69bdae19d6ebd42d80a30a8ce5818fee7944a0c6e7310504066f54562d4c494e2d382d706f6c696379075840918c7bc5ec775e6cf0ce934a20e18f38cb54c3e3e7f8a3c9c29a1871c9257a5e1358fac9e16f51aae8a54ca077a258f1c6b37b82b9109836798f3b6254eb5fec0802095840778db69207a3d6dc56b6a12adc3a9923e75b3c4c5f9ec2aa8a73068a16cdebc998f8ece08d4726e24cfcf0955e03a37dfa1672503843e5617d78552078a2d2ae`
- `tags_hash`: `ee8446f14a548be347b458aa56f2751af5640edf24408b325884a2ca8c65cb579ae5471159329298fc793222304f40ed51aa9b940ca6bb6ccae41cdc4e4eb0e3`
- `policy_compiled`: `a500010104020203828458200100000000000000000000000000000000000000000000000000000000000000582001000000000000000000000000000000000000000000000000000000000000005820000000000000000000000000000000000000000000000000000000000000000058200000000000000000000000000000000000000000000000000000000000000000845820000000000000000000000000000000000000000000000000000000000000000058200000000000000000000000000000000000000000000000000000000000000000582003000000000000000000000000000000000000000000000000000000000000005820ecd3f55c1a631258d69cf7a2def9de140000000000000000000000000000001004825820e80300000000000000000000000000000000000000000000000000000000000058200700000000000000000000000000000000000000000000000000000000000000`
- `policy_hash`: `918c7bc5ec775e6cf0ce934a20e18f38cb54c3e3e7f8a3c9c29a1871c9257a5e1358fac9e16f51aae8a54ca077a258f1c6b37b82b9109836798f3b6254eb5fec`

Base encodings:

- `G_pol`: `189e14198a4e5dfb141b576a48541cf410eb808c2d718eeaef8fea12e7efb047`
- `H_pol`: `fc84a1de9950310234689fadf41b0eadc5fadc708217204d2045cf92670cc75a`
- $B^r_1,\ldots,B^r_4$:
  1. `f84fd7c6f34981399e4bfdbdcfc429e17d108fc674a05422865a30e2563e871a`
  2. `88fd90f0826cff3290bfeeb6211763d28c8c4a8416089b8c5921e35d0554fe10`
  3. `644f59a9ed0a7eb7f3cf690c7f540e1e05e71bf36c41398c357b7441310d1a69`
  4. `48da174cf81b9dffc80f35ce968a5e9f2e1418038454900fa9359105656aa134`
- $B^m_1,\ldots,B^m_4$:
  1. `3c211f263f15efaadd19c52166503a384ba46933913df0b7e3fbecc2ebce2d42`
  2. `e4486231b0d8ebc2579bad1c9959b3fab62655e701f840de22494ac65b122529`
  3. `66e096a992511c7725592f12095eac7ada0a363af424bf3ba9e266ba78a9e267`
  4. `fc2bdb30a58b307a0a5aef4a3841beebe25dce086763ac49fe425570ff2ce447`

Commitments $C_1,\ldots,C_8$:

1. `603b00e71566e8994fe80d73d5fa7e39419484b13228e22a8bbc150c6e8db214`
2. `84c3e95092821c4079cc737fc5317273e1203be7dfec73ccde81aabb50044111`
3. `76730553d30019f348641399720866c50812cd0491109b5d47b843a8d19e234c`
4. `acc8895c27d8b147d9aa076e29884d3c940e4be5cfab17e03569fad42c3ac960`
5. `c656cc856d94d7ec014470d8e6b423064a0feec93699803f2c38e943ff074f76`
6. `da094c62b00a1a9cb9982295033401b34a11c5628dacf6c28fda9f4329f52f7e`
7. `8c4532bda29fa683dea8891e7791149adcd72a7cbad60dcc5d656329c9e6666f`
8. `4e6b8edb2ba516f83acc20c52203482281706f85b9ff9a3ad154c125db6e9e05`

Transcript roots $R_0,\ldots,R_8$:

- $R_0$: `c976527b7f92a96a09238428b2584310960831e0494b135285cf38bfaddaab5f15a13b9ac00154d1a44aa72efe22eaadaea4dcc3d9f51e4e7655f22c5d02eac1`
- $R_1$: `de541818200bfcd65967c8392dbc8880bd8ef4d415738e353ecaec7c75f85b5c73ac9aaabaf6deec791d714357487485591f871e5b0a7656a71876d46c602eb1`
- $R_2$: `7dc5eca7adce9331ed424e3fc55d1495710b420ebc8b1bb2b368efa53d6618f815c427c3ef5c16a76ca621ec969f809171b9797c66b3f296ce329d67254e86e9`
- $R_3$: `d71b5d30e7d95588caf5cf8a28fd3c6ec8d1d325bb6cfa23532249bd4413734edfcbe40b7236d8399583bc92e4e77ed7b8f29f5e9fa0c379e12ada6ce55d0468`
- $R_4$: `1bcdca39dd6a017d481fd51f1e69b5c4f43da990867e66e9b7b92de81799109e8cbee60fd0ca57eb445f92a5cd6f98b8eb090bf29b17fedfb5c5b88ce7932e5a`
- $R_5$: `224dd25e07325792bef0fb7d88224213bdd6572c767c491a60fb788a5a924ec493b7cff158223a5821a9deffa8f04d3e218d1fdaf7158777f9147a2c98b978aa`
- $R_6$: `48329eac48fb2a10d293612cc0edf6444128ce6356f20650483a278d103bcd5e45b130969050ba48671518cdccd6a3d9e7aacfe1266df5ae918d3837ba4efe25`
- $R_7$: `38ec7f9a1e93139bfdaa4a699c083ad0cca33df26003a47ebae2354085a7baf70291f3d848b2983cec9702fa39596afbc0bc9f0015f7bd31bc020bac02c76748`
- $R_8$: `cbc702ee275a756538ebc7e90d9c18686efb0adc6545623a2b0aef5f011ceadbfeb9769ae7dedfd18cc0e068eceaa8d92a9accfbcad6d066cf3bfb4be1b1bab5`

Challenges $\alpha_1,\ldots,\alpha_8$:

- $\alpha_1$: `b48e54cd31eaa9af43a26b00d7fc2cf08bb97cd35dfc3f1e1df09c85a0148703`
- $\alpha_2$: `4fe5d12ced554d9434efde6ea0885cd6da4a5bcf537aa864c2d4cecfdeeeee08`
- $\alpha_3$: `75ae7e0c6433cf46c2490e22c1b20ca782c63eebb11bcf6b0ad1f76c85969e0b`
- $\alpha_4$: `34a72b9d8e59e6a7061dcced884af98bf0ea6941933f8bbb1b7302c3c2f0090b`
- $\alpha_5$: `6e064e775017a1fd839d19301dfe511a0246963d8525fbdde8d061a8e0df160d`
- $\alpha_6$: `260d59aa6e3bc09edb7076c7f6c74456be977b6f7119771c0ac4f5dbe366a805`
- $\alpha_7$: `b56102cb5ad0e19c6f636056780a8987b382a79b6ef4d552a599d198ab776007`
- $\alpha_8$: `6e3645511708914e449068f869d42ca40ff2723a063c1c68e6200c09e4c7e302`

Folded commitment and policy commitments:

- $C_\star$: `e418f98b8c0140f533664e8c4526876c499f76088d67d3bf8af928b95054b643`
- $V_1,\ldots,V_4$:
  - $V_1$: `fcc9bdcc2d15286173d3008516e802b90180a187fc7128a5a512b75e03445d5d`
  - $V_2$: `b2d4e1b5da8ff40533d89e7cc7bb4f0a7bc0e041c7c2e002072b25c8ecbbcd6d`
  - $V_3$: `5a6e65add337638724761264b090b12d30d68c1da9641aae88b96e6a5333fa48`
  - $V_4$: `2250e483b032393ba955a822c8962f4904af701dd9359903a0ddde80cb4d0876`

Proof $\pi_{\mathrm{link}}$:

- `T_C`: `563ffae71ce94b46a650be9083ca3d07e6d9899f0e2411693b29854a38f7b70c`
- `T_V`: `['8e79e00b09fefd6115202ff8bf81f73bb17057d2c3b888b95db35eab7bf23f57', 'c8baf90c187b3eed3f7f8f98497d5105eea8672dab4db1571ca047cad6794c0c', 'b2bf518e8d498df943f0652ee527e48c93a2a7cca2134129613a802f68a9a436', '082be7d54cd6186b1d817ce183f3aeceb1d7297f80125847ecf5ac8dc75d1075']`
- `z_r`: `['aca6f4a10e0db4524166d53eafd96e50bde88e3e6c58ad88c1ca66b127202e0d', '85e35d6b39e865eae49b2ab55aebb7b94b73ce7b1ed22e0f75bf90e17b10cc08', '2b5ec711b9b7553cfaf9223fd4ac6c57ac1d70aa2e37ceddb2fc7f2731c2730c', 'a36535af31f5f14e0ac33968b0a75506eed8182607fb21aa15d4429d06cb1d0a']`
- `z_m`: `['0a4ea2fea7859c30657fc765d924212445896a896adde5c25cfab9bafbd28905', 'd8db8753237e4d18dfe5a7970d527aaaf624d661d1dbbfcfa857e0be074f1903', '9b349b57f8a11803a660382a39d378b03e7464ccfa95d09ee56fbda97b223c03', 'ba06edb44f58a35c0be88cc2f7879348b34b31c6afa0e423f327c5e1ffabb40a']`
- `z_gamma`: `['820d4a42c04d9874350b762151889e1051c817bd8523d387a371a4f1a8d1d902', 'f80264bb338a38697f5b70961dda671adec2e15ba73322bc16104477d743fb0a', '7c233927e5e3e5b545c42173bcc7991b9101bfff94370508375e8a9863e25a07', 'd929eb80df910db824b32205a8390ad936613be9f73f33b4d9cba12b1c7e0907']`
- `c_link`: `d6d4df7f8e5cc18151cf81d2de892a064c6499de9624df650ad660ae5bcf3f09`

Proof $\pi_{\mathrm{cons}}$:

- $W$: `be81a652279bb04460b748f080c3352ad1ae23a32fa6685ca07f96f931edd02d`
- $T$: `808e093465c0309f776561624f6551f3480b0a36e89d777e87e32803f75c450f`
- $z$: `1d4346d0f9ae19bc2dde410f49f404e429cc1c06f95208496370d9dfd342b70f`
- `c_cons`: `8fd414fce0639090f86deb0c548084d800bad003e30a47170f97369d7bcbd20f`

JSON bundle:

```json
{
  "name": "TV-LIN-8",
  "suite_id": "NESSA-EC-RISTRETTO255-SHA512-v1",
  "d": 4,
  "k_rows": 2,
  "N": 8,
  "deterministic_seed_hex": "4e455353412d45433a746573742d766563746f723a54562d4c494e2d383a7631",
  "tags_hex": "aa000101781f4e455353412d45432d52495354524554544f3235352d5348413531322d763102782472697374726574746f3235355f584d443a5348412d3531325f523235354d41505f524f5f036854562d4c494e2d38045840287034d3d090faf7a57d7c4a812225e8026fb12e6c08a0f1e07dc4dc0977d2f43da658add3b4ecff118f69bdae19d6ebd42d80a30a8ce5818fee7944a0c6e7310504066f54562d4c494e2d382d706f6c696379075840918c7bc5ec775e6cf0ce934a20e18f38cb54c3e3e7f8a3c9c29a1871c9257a5e1358fac9e16f51aae8a54ca077a258f1c6b37b82b9109836798f3b6254eb5fec0802095840778db69207a3d6dc56b6a12adc3a9923e75b3c4c5f9ec2aa8a73068a16cdebc998f8ece08d4726e24cfcf0955e03a37dfa1672503843e5617d78552078a2d2ae",
  "tags_hash_hex": "ee8446f14a548be347b458aa56f2751af5640edf24408b325884a2ca8c65cb579ae5471159329298fc793222304f40ed51aa9b940ca6bb6ccae41cdc4e4eb0e3",
  "policy_rows": [
    [
      1,
      1,
      0,
      0
    ],
    [
      0,
      0,
      3,
      -1
    ]
  ],
  "policy_targets": [
    1000,
    7
  ],
  "policy_compiled_hex": "a500010104020203828458200100000000000000000000000000000000000000000000000000000000000000582001000000000000000000000000000000000000000000000000000000000000005820000000000000000000000000000000000000000000000000000000000000000058200000000000000000000000000000000000000000000000000000000000000000845820000000000000000000000000000000000000000000000000000000000000000058200000000000000000000000000000000000000000000000000000000000000000582003000000000000000000000000000000000000000000000000000000000000005820ecd3f55c1a631258d69cf7a2def9de140000000000000000000000000000001004825820e80300000000000000000000000000000000000000000000000000000000000058200700000000000000000000000000000000000000000000000000000000000000",
  "policy_hash_hex": "918c7bc5ec775e6cf0ce934a20e18f38cb54c3e3e7f8a3c9c29a1871c9257a5e1358fac9e16f51aae8a54ca077a258f1c6b37b82b9109836798f3b6254eb5fec",
  "compressed_policy_hash_hex": "918c7bc5ec775e6cf0ce934a20e18f38cb54c3e3e7f8a3c9c29a1871c9257a5e1358fac9e16f51aae8a54ca077a258f1c6b37b82b9109836798f3b6254eb5fec",
  "compressed_coeffs_hex": [
    "1800e1b666b2c05e12c7f7098e2faf7d540022c7b42dd494e61559145502dc08",
    "1800e1b666b2c05e12c7f7098e2faf7d540022c7b42dd494e61559145502dc08",
    "14297d6f2f1883eedfc48e304a3be305c05c83de96ca840cf6facf28ef77a405",
    "f338284f4e6edacdfc47787b31ea535a15e17e6078bcd3fbad01109d05d87303"
  ],
  "compressed_target_hex": "58fbd49ca50ecb483351a46f89995fea9179b65190aaaff2f1b93de33d654807",
  "generators": {
    "Br": [
      "f84fd7c6f34981399e4bfdbdcfc429e17d108fc674a05422865a30e2563e871a",
      "88fd90f0826cff3290bfeeb6211763d28c8c4a8416089b8c5921e35d0554fe10",
      "644f59a9ed0a7eb7f3cf690c7f540e1e05e71bf36c41398c357b7441310d1a69",
      "48da174cf81b9dffc80f35ce968a5e9f2e1418038454900fa9359105656aa134"
    ],
    "Bm": [
      "3c211f263f15efaadd19c52166503a384ba46933913df0b7e3fbecc2ebce2d42",
      "e4486231b0d8ebc2579bad1c9959b3fab62655e701f840de22494ac65b122529",
      "66e096a992511c7725592f12095eac7ada0a363af424bf3ba9e266ba78a9e267",
      "fc2bdb30a58b307a0a5aef4a3841beebe25dce086763ac49fe425570ff2ce447"
    ],
    "G_pol": "189e14198a4e5dfb141b576a48541cf410eb808c2d718eeaef8fea12e7efb047",
    "H_pol": "fc84a1de9950310234689fadf41b0eadc5fadc708217204d2045cf92670cc75a"
  },
  "events": [
    {
      "index": 1,
      "values": [
        100,
        900,
        13,
        32
      ],
      "m_hex": [
        "6400000000000000000000000000000000000000000000000000000000000000",
        "8403000000000000000000000000000000000000000000000000000000000000",
        "0d00000000000000000000000000000000000000000000000000000000000000",
        "2000000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "a28fdd4b1df91f8871baaf10da5bbb72d82875ff8c73cb10471e54840e31a70a",
        "c88f5242e570ab0cd062bc2b07b5665b7e7d22f9e50ec574129c253a0e2ce70e",
        "6dc3e5128596fa0bc516c3b348b608a832257b10648314bfe9d8e3c90d4f0d09",
        "12b289dc250b46d2eb1e837db356e7309ec6b212f0a49f7a0a631917ebbb4a0a"
      ],
      "commitment_hex": "603b00e71566e8994fe80d73d5fa7e39419484b13228e22a8bbc150c6e8db214"
    },
    {
      "index": 2,
      "values": [
        200,
        800,
        23,
        62
      ],
      "m_hex": [
        "c800000000000000000000000000000000000000000000000000000000000000",
        "2003000000000000000000000000000000000000000000000000000000000000",
        "1700000000000000000000000000000000000000000000000000000000000000",
        "3e00000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "069d707d3627703988083278af5b3149bf4a687e005d8170e080ae4aa938d306",
        "80fe60e94b97c059f2f8286b2b459aadb20c212325aa0660db768e213bdbf602",
        "a790c45e495d55b45511965fac7e51c62486885259f6394aeeb37541438cc109",
        "172aec965ce2f81065bd9b9ba62152e74b0aa6bf7a0d938844d35c5afa3fb706"
      ],
      "commitment_hex": "84c3e95092821c4079cc737fc5317273e1203be7dfec73ccde81aabb50044111"
    },
    {
      "index": 3,
      "values": [
        300,
        700,
        33,
        92
      ],
      "m_hex": [
        "2c01000000000000000000000000000000000000000000000000000000000000",
        "bc02000000000000000000000000000000000000000000000000000000000000",
        "2100000000000000000000000000000000000000000000000000000000000000",
        "5c00000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "6f214868ffdc8e381ff486eb5419c953084a7cc46713b85f57cc48339979cd03",
        "ee293d3af4112ca5b751c5133b91b526a281cc82710934d9c669f293eb586c0c",
        "ac73046b2e52c9ce40e009b99b1604453f2ccf8dd03b3d5eb5d21339c1c5e109",
        "41eb7b09f89220004ee02281b29323b6c0e25dae79cd9447f14189fe75b72909"
      ],
      "commitment_hex": "76730553d30019f348641399720866c50812cd0491109b5d47b843a8d19e234c"
    },
    {
      "index": 4,
      "values": [
        400,
        600,
        43,
        122
      ],
      "m_hex": [
        "9001000000000000000000000000000000000000000000000000000000000000",
        "5802000000000000000000000000000000000000000000000000000000000000",
        "2b00000000000000000000000000000000000000000000000000000000000000",
        "7a00000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "58260153cd81d09899d24de2583cd278389a28e5f955101e7c9c13e664c60b02",
        "c42a2e1d0343860d7420ce7a3662e7e5d8275643d8e94513a15dc1a101499603",
        "88ebd0f3c6df630358b2917a0afac4f7516629f6c8fbabc72a40a0db7376b803",
        "d934d6e2fa5e8a5b5c1563c001914a47df9f5335129a671b059a181098faf20d"
      ],
      "commitment_hex": "acc8895c27d8b147d9aa076e29884d3c940e4be5cfab17e03569fad42c3ac960"
    },
    {
      "index": 5,
      "values": [
        500,
        500,
        53,
        152
      ],
      "m_hex": [
        "f401000000000000000000000000000000000000000000000000000000000000",
        "f401000000000000000000000000000000000000000000000000000000000000",
        "3500000000000000000000000000000000000000000000000000000000000000",
        "9800000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "7462c027eeb4eaf1e06ae1f7b7b8cfc86db8328afd32271b4a6ab2df71c92603",
        "d1b339b4925ec9135af8057c9b6e6a1712c7b3e3e31ad4b4de8abb90c1eefc0c",
        "c5bf6f3dc9350cff560d050c0e208c2fb8eb6a3bbb63f7a1f7be61424983800f",
        "42bbca0d3177b22402c367c8f35ed86db25129f765dad8a8829c39d85a7e3b00"
      ],
      "commitment_hex": "c656cc856d94d7ec014470d8e6b423064a0feec93699803f2c38e943ff074f76"
    },
    {
      "index": 6,
      "values": [
        600,
        400,
        63,
        182
      ],
      "m_hex": [
        "5802000000000000000000000000000000000000000000000000000000000000",
        "9001000000000000000000000000000000000000000000000000000000000000",
        "3f00000000000000000000000000000000000000000000000000000000000000",
        "b600000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "788d60ec9c234f41dfe902c56575a8d61747120b5c8c709f14ed956f59beca0c",
        "1451e069af5c095cec1598e5181ffde81aaf512184c74e6e73da73e817c1550e",
        "1fd34afde5788664b004762982e0ac32e802b12bfd6e3cfb304e94e5c0de4506",
        "1a0919828423174830177daba316c55f15ecafcbac6ae1f3e979ad4fcff0ec09"
      ],
      "commitment_hex": "da094c62b00a1a9cb9982295033401b34a11c5628dacf6c28fda9f4329f52f7e"
    },
    {
      "index": 7,
      "values": [
        700,
        300,
        73,
        212
      ],
      "m_hex": [
        "bc02000000000000000000000000000000000000000000000000000000000000",
        "2c01000000000000000000000000000000000000000000000000000000000000",
        "4900000000000000000000000000000000000000000000000000000000000000",
        "d400000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "67f4cd0dd4b1dc6829499aa30fda75f846669edcf305ddaf45e3638cac9c230c",
        "cf6afe52bf412fba1c9cfb1d1b8348291f37cd63a4600bb80f4d6185831da70e",
        "596e6e05fa91ec7cbc119852f508bc6abb1c1d3adc56057f41040d1f6725c208",
        "310434c0103fb82a3447e2c81ff7e66fe5384ac1244dc6cdd32583df06b1e50b"
      ],
      "commitment_hex": "8c4532bda29fa683dea8891e7791149adcd72a7cbad60dcc5d656329c9e6666f"
    },
    {
      "index": 8,
      "values": [
        800,
        200,
        83,
        242
      ],
      "m_hex": [
        "2003000000000000000000000000000000000000000000000000000000000000",
        "c800000000000000000000000000000000000000000000000000000000000000",
        "5300000000000000000000000000000000000000000000000000000000000000",
        "f200000000000000000000000000000000000000000000000000000000000000"
      ],
      "rho_hex": [
        "c68eb3ca4dac2cbe7aef984b00d195845538ff6cd5f859a6c4c5a1316c75990f",
        "6cd4d9a78bfe28efda3837506200a63167f66413aef020119e30e45d92e4ed0d",
        "41a238cfa8a76a7866707631f28fdbaa9f6bda7bd8b66958e3bc33442c467f07",
        "01d6eeca441fe9ff45d416c23b0acb6911ef04514fcd44d7542649b3887a310e"
      ],
      "commitment_hex": "4e6b8edb2ba516f83acc20c52203482281706f85b9ff9a3ad154c125db6e9e05"
    }
  ],
  "transcript": {
    "R_hex": [
      "c976527b7f92a96a09238428b2584310960831e0494b135285cf38bfaddaab5f15a13b9ac00154d1a44aa72efe22eaadaea4dcc3d9f51e4e7655f22c5d02eac1",
      "de541818200bfcd65967c8392dbc8880bd8ef4d415738e353ecaec7c75f85b5c73ac9aaabaf6deec791d714357487485591f871e5b0a7656a71876d46c602eb1",
      "7dc5eca7adce9331ed424e3fc55d1495710b420ebc8b1bb2b368efa53d6618f815c427c3ef5c16a76ca621ec969f809171b9797c66b3f296ce329d67254e86e9",
      "d71b5d30e7d95588caf5cf8a28fd3c6ec8d1d325bb6cfa23532249bd4413734edfcbe40b7236d8399583bc92e4e77ed7b8f29f5e9fa0c379e12ada6ce55d0468",
      "1bcdca39dd6a017d481fd51f1e69b5c4f43da990867e66e9b7b92de81799109e8cbee60fd0ca57eb445f92a5cd6f98b8eb090bf29b17fedfb5c5b88ce7932e5a",
      "224dd25e07325792bef0fb7d88224213bdd6572c767c491a60fb788a5a924ec493b7cff158223a5821a9deffa8f04d3e218d1fdaf7158777f9147a2c98b978aa",
      "48329eac48fb2a10d293612cc0edf6444128ce6356f20650483a278d103bcd5e45b130969050ba48671518cdccd6a3d9e7aacfe1266df5ae918d3837ba4efe25",
      "38ec7f9a1e93139bfdaa4a699c083ad0cca33df26003a47ebae2354085a7baf70291f3d848b2983cec9702fa39596afbc0bc9f0015f7bd31bc020bac02c76748",
      "cbc702ee275a756538ebc7e90d9c18686efb0adc6545623a2b0aef5f011ceadbfeb9769ae7dedfd18cc0e068eceaa8d92a9accfbcad6d066cf3bfb4be1b1bab5"
    ],
    "R_final_hex": "cbc702ee275a756538ebc7e90d9c18686efb0adc6545623a2b0aef5f011ceadbfeb9769ae7dedfd18cc0e068eceaa8d92a9accfbcad6d066cf3bfb4be1b1bab5",
    "alphas_hex_le": [
      "b48e54cd31eaa9af43a26b00d7fc2cf08bb97cd35dfc3f1e1df09c85a0148703",
      "4fe5d12ced554d9434efde6ea0885cd6da4a5bcf537aa864c2d4cecfdeeeee08",
      "75ae7e0c6433cf46c2490e22c1b20ca782c63eebb11bcf6b0ad1f76c85969e0b",
      "34a72b9d8e59e6a7061dcced884af98bf0ea6941933f8bbb1b7302c3c2f0090b",
      "6e064e775017a1fd839d19301dfe511a0246963d8525fbdde8d061a8e0df160d",
      "260d59aa6e3bc09edb7076c7f6c74456be977b6f7119771c0ac4f5dbe366a805",
      "b56102cb5ad0e19c6f636056780a8987b382a79b6ef4d552a599d198ab776007",
      "6e3645511708914e449068f869d42ca40ff2723a063c1c68e6200c09e4c7e302"
    ]
  },
  "folded": {
    "C_star_hex": "e418f98b8c0140f533664e8c4526876c499f76088d67d3bf8af928b95054b643",
    "m_star_hex": [
      "417cba06e0b034bad3400565b6eaa83009e48457b8ccd2f36145377b198cc50f",
      "dceb5fc1bce2c74528bb5c3d5c2f398541c8639c5a9b84b1b37f9adbfc474d05",
      "0a0146df2fede7667283740886bb3a237fc9fa33063ff1507d9057c288f4f909",
      "6827833f83712165a13c09e306772f84f021355962f34055d945c795d865ff0c"
    ],
    "rho_star_hex": [
      "40c8295fb933c0de881785bc4cd865560133dda18cbeb761f9e3c1023fae4502",
      "a2609ced8ee33527b24d4a03b4f6c40b15c07e111f127f9300f8b287237d780d",
      "dcf76708c7b78553f19f9406425eae9fa8166aaa44511f8f9ab56a927e58cf0c",
      "eae94887df4b0e5dca376275a89ab44c99c6335807c65ca8964a5080b93d9001"
    ]
  },
  "policy_commitments_hex": [
    "fcc9bdcc2d15286173d3008516e802b90180a187fc7128a5a512b75e03445d5d",
    "b2d4e1b5da8ff40533d89e7cc7bb4f0a7bc0e041c7c2e002072b25c8ecbbcd6d",
    "5a6e65add337638724761264b090b12d30d68c1da9641aae88b96e6a5333fa48",
    "2250e483b032393ba955a822c8962f4904af701dd9359903a0ddde80cb4d0876"
  ],
  "pi_link": {
    "T_C_hex": "563ffae71ce94b46a650be9083ca3d07e6d9899f0e2411693b29854a38f7b70c",
    "T_V_hex": [
      "8e79e00b09fefd6115202ff8bf81f73bb17057d2c3b888b95db35eab7bf23f57",
      "c8baf90c187b3eed3f7f8f98497d5105eea8672dab4db1571ca047cad6794c0c",
      "b2bf518e8d498df943f0652ee527e48c93a2a7cca2134129613a802f68a9a436",
      "082be7d54cd6186b1d817ce183f3aeceb1d7297f80125847ecf5ac8dc75d1075"
    ],
    "z_r_hex": [
      "aca6f4a10e0db4524166d53eafd96e50bde88e3e6c58ad88c1ca66b127202e0d",
      "85e35d6b39e865eae49b2ab55aebb7b94b73ce7b1ed22e0f75bf90e17b10cc08",
      "2b5ec711b9b7553cfaf9223fd4ac6c57ac1d70aa2e37ceddb2fc7f2731c2730c",
      "a36535af31f5f14e0ac33968b0a75506eed8182607fb21aa15d4429d06cb1d0a"
    ],
    "z_m_hex": [
      "0a4ea2fea7859c30657fc765d924212445896a896adde5c25cfab9bafbd28905",
      "d8db8753237e4d18dfe5a7970d527aaaf624d661d1dbbfcfa857e0be074f1903",
      "9b349b57f8a11803a660382a39d378b03e7464ccfa95d09ee56fbda97b223c03",
      "ba06edb44f58a35c0be88cc2f7879348b34b31c6afa0e423f327c5e1ffabb40a"
    ],
    "z_gamma_hex": [
      "820d4a42c04d9874350b762151889e1051c817bd8523d387a371a4f1a8d1d902",
      "f80264bb338a38697f5b70961dda671adec2e15ba73322bc16104477d743fb0a",
      "7c233927e5e3e5b545c42173bcc7991b9101bfff94370508375e8a9863e25a07",
      "d929eb80df910db824b32205a8390ad936613be9f73f33b4d9cba12b1c7e0907"
    ],
    "challenge_hex": "d6d4df7f8e5cc18151cf81d2de892a064c6499de9624df650ad660ae5bcf3f09"
  },
  "pi_cons": {
    "W_hex": "be81a652279bb04460b748f080c3352ad1ae23a32fa6685ca07f96f931edd02d",
    "T_hex": "808e093465c0309f776561624f6551f3480b0a36e89d777e87e32803f75c450f",
    "z_hex": "1d4346d0f9ae19bc2dde410f49f404e429cc1c06f95208496370d9dfd342b70f",
    "challenge_hex": "8fd414fce0639090f86deb0c548084d800bad003e30a47170f97369d7bcbd20f"
  },
  "verification": {
    "link_verify_ok": true,
    "cons_verify_ok": true,
    "proof_size_bytes": 832
  }
}
```

Verification summary:

- `link_verify_ok = True`
- `cons_verify_ok = True`
- `proof_size_bytes = 832`

#### TV-R1CS-8: Non-linear Folding Example (Multiplication Gate)

Parameters:

- $d = 3$
- $N = 8$

Per-event values:

- $e_1$: $x = 2$, $y = 5$, $z = 10$
- $e_2$: $x = 5$, $y = 12$, $z = 60$
- $e_3$: $x = 8$, $y = 19$, $z = 152$
- $e_4$: $x = 11$, $y = 26$, $z = 286$
- $e_5$: $x = 14$, $y = 33$, $z = 462$
- $e_6$: $x = 17$, $y = 40$, $z = 680$
- $e_7$: $x = 20$, $y = 47$, $z = 940$
- $e_8$: $x = 23$, $y = 54$, $z = 1242$

Folded values:

- $x_\star$ = `c3bd5d9698764d0fb66d4c2bebeee43a9a2285e955230682ae4ca08fafef260e`
- $y_\star$ = `41b465c89be253bb6bc7dd2cc74fe5c25903c1e63d200531fd377a70a3cca502`
- $z_\star$ = `04785385d765ce2a34454554ced9515cf645c0b90ae76537538d40fe75590b06`
- $E$ = `9964587ac4e063de167517cc7263134c5a2c648e88bbd57a06261aef9a92b804`
- Identity: $x_\star y_\star \equiv z_\star + E \pmod L$
- Identity holds: `True`

JSON bundle:

```json
{
  "name": "TV-R1CS-8",
  "suite_id": "NESSA-EC-RISTRETTO255-SHA512-v1",
  "d": 3,
  "N": 8,
  "deterministic_seed_hex": "4e455353412d45433a746573742d766563746f723a54562d523143532d383a7631",
  "events": [
    {
      "index": 1,
      "x": 2,
      "y": 5,
      "z": 10,
      "z_equals_xy": true,
      "commitment_hex": "6c536280eb1234309ae2292a742ebd55fd31d3893c2238165ff7e65ee3de7517"
    },
    {
      "index": 2,
      "x": 5,
      "y": 12,
      "z": 60,
      "z_equals_xy": true,
      "commitment_hex": "d4059976a16f5f51826eabe75b18644cee96e9c36c772867fc77eddb24e0e473"
    },
    {
      "index": 3,
      "x": 8,
      "y": 19,
      "z": 152,
      "z_equals_xy": true,
      "commitment_hex": "c2bee75fa743fe4003f707d6d76b52ad16ff85d187735efcf4ae79355f992c1c"
    },
    {
      "index": 4,
      "x": 11,
      "y": 26,
      "z": 286,
      "z_equals_xy": true,
      "commitment_hex": "606a5ce79cc7ea1f0346ddeb705197d0a4676de622e190cfed182bd29a285425"
    },
    {
      "index": 5,
      "x": 14,
      "y": 33,
      "z": 462,
      "z_equals_xy": true,
      "commitment_hex": "8eb85e5674f9f59db4d0643e2805a9f0ed3b17d4e4f05292a69c5dab35465169"
    },
    {
      "index": 6,
      "x": 17,
      "y": 40,
      "z": 680,
      "z_equals_xy": true,
      "commitment_hex": "20c5ee26d14a86ca88fc998f9f3c8282b607f4ade3a24af659d3cef6549b2911"
    },
    {
      "index": 7,
      "x": 20,
      "y": 47,
      "z": 940,
      "z_equals_xy": true,
      "commitment_hex": "d4e49c31e9b70dbf70d2a1543769dcb2ddb8c342dad4fe30411152075b016e1b"
    },
    {
      "index": 8,
      "x": 23,
      "y": 54,
      "z": 1242,
      "z_equals_xy": true,
      "commitment_hex": "f4126fe1da50e85d1634675be1f4ed87f7ac2ba0342ad1f9253ef6f89abc4570"
    }
  ],
  "tags_hash_hex": "9dd7d7b6abb113d072b013a0cd1bd8418b33cd1c29ed7df1d4e9dacdf93433ce6c34833d22b894800e1afabaf284da108cb12cabeb4536b01f3b0d835acfd6e3",
  "transcript_roots_hex": [
    "64ecc270fd900f574b538f8896c5cd0a14e3d28df6a32f39edfc55a3d698bbd3c1c9cc24c8917b96bf0159cdb3d12a9ca22620cca3b94d4436e5406196086ae4",
    "290ee679df75e027db0ece422bd28232c5170a9e493e5842e8cfb6bfa184c26ecda6896d15980d2297337c270f76108099feddb9bd3d44243eaa0cbc42e741ea",
    "e10df567ac0fa72d80f1a7c14bdbb6cbdc698f7ee29529f82e56eed73e9ce08cf55eefd8aded4fbb9875d34b0a5e4a571f788cc829eb0a230e4fb6023da172cf",
    "7307dbd127dca1d1eebe10f24ae1b92b2917fba175f4b1792953ef1379f052f726db1b1107d9327d2d6c0e9a6e47fa9978f9bc7c499c9eccf1087138e529fe08",
    "b8c25bcf36a2646ad643b6f54db58386724cfa01dab1c057f11e5047d32cce82ca1be78dd0257783d36f0088cb8310e3c82c5e43f2f8ba0c41450f5bd381db5f",
    "4826fa916b2a05266eabec52f061435c04655ec484ce3ee71784a8576080493bebde28358ea213d525f7f7cb855a21904a148b37d5bd3ac411e041e3e7ca4f60",
    "223f2b1c21780fc4cf56dda466e16eba09861a01f54f36335f55067913fef841dcbac882979a90eac01ac36238f97b09f18038274ed14558d2c593947b885d37",
    "da67d01e16ddc38ece15c6d5e5eb853e1197e02c8844ef059d131debe7b62d9cd72eaab2fd04569a3f3be0d53b8b43ca196ffde8cb4951ec1c269eaa3344499b",
    "e8ea7acfe4fe73549126ede24435bae42a39c0121812f7b8f9e9648155f02b273568f563904396d70738c8df561ade043221c37a00a5fc99896b0a0e015cf7fa"
  ],
  "R_final_hex": "e8ea7acfe4fe73549126ede24435bae42a39c0121812f7b8f9e9648155f02b273568f563904396d70738c8df561ade043221c37a00a5fc99896b0a0e015cf7fa",
  "alphas_hex_le": [
    "ab8c34feffc3053671851c75c36c8808b1fa4e6d6da1dde363e0edc60bc4a00d",
    "e7ed54c7126168ed2bed03e9c318909cc76a160bad1d40bcfd75c01cacaa5b0c",
    "c31d6f03ffe7df7d7e8b2a99dac5a4fbc0f6098c028a03d93c253e9a42518300",
    "9c022fab56179415980c1f0beb49d248ff5d0a377a9d355ed0fd31c9f077400f",
    "5eb9bc6114a22723096155dbb03e07813f048be526f540a35d0cba7edcf26200",
    "a1a45926f9cdd083efa0a903cca33b549ea6e47fb6a514130d16dbe3eb7a9404",
    "8f55bc4307881ef78f0253aaecbe68da4a01cdf08c02c68f9df14346ce323f0d",
    "31e5409e312c9ce26b677229e3f2e7e374b1e8bf5ee571e7ba0115749bffe908"
  ],
  "folded": {
    "C_star_hex": "8026a271aa13ef3312459304703b7060fefaf10f530547b99ad1ef307f04dd4b",
    "x_star": 6401174319120458851449634732545793179196184414240138391883149417863819673027,
    "y_star": 1197567825426272738472658092656015811773010697951104382603076729860472026177,
    "z_star": 2733929845726662232020004572730806956764090439519530187701655305576315582468,
    "x_star_hex": "c3bd5d9698764d0fb66d4c2bebeee43a9a2285e955230682ae4ca08fafef260e",
    "y_star_hex": "41b465c89be253bb6bc7dd2cc74fe5c25903c1e63d200531fd377a70a3cca502",
    "z_star_hex": "04785385d765ce2a34454554ced9515cf645c0b90ae76537538d40fe75590b06",
    "E": 2135363086231101026376067755242140314090524063141191156100925275794438644889,
    "E_hex": "9964587ac4e063de167517cc7263134c5a2c648e88bbd57a06261aef9a92b804",
    "identity": "x_star * y_star = z_star + E (mod L)",
    "identity_holds": true
  },
  "generators": {
    "Br": [
      "f84fd7c6f34981399e4bfdbdcfc429e17d108fc674a05422865a30e2563e871a",
      "88fd90f0826cff3290bfeeb6211763d28c8c4a8416089b8c5921e35d0554fe10",
      "644f59a9ed0a7eb7f3cf690c7f540e1e05e71bf36c41398c357b7441310d1a69"
    ],
    "Bm": [
      "3c211f263f15efaadd19c52166503a384ba46933913df0b7e3fbecc2ebce2d42",
      "e4486231b0d8ebc2579bad1c9959b3fab62655e701f840de22494ac65b122529",
      "66e096a992511c7725592f12095eac7ada0a363af424bf3ba9e266ba78a9e267"
    ],
    "G_pol": "189e14198a4e5dfb141b576a48541cf410eb808c2d718eeaef8fea12e7efb047",
    "H_pol": "fc84a1de9950310234689fadf41b0eadc5fadc708217204d2045cf92670cc75a"
  },
  "verification": {
    "folding_check_ok": true,
    "link_verify_ok": true,
    "cons_verify_ok": true,
    "proof_size_bytes": 1120
  }
}
```

## Required Protocol Decisions / Open Items

Item: Hidden count/order profile
Why it matters: $\alpha_i$ derivation requires an index domain; hiding $N$ and/or order requires padding or redesign.
Exact decision needed: Decide whether v1 leakage ($N$/order visible) is acceptable or whether a padded length $L$ is required as default.
Allowed options: (a) $N$/order public (v1); (b) padded $L$ with dummy events and defined dummy semantics; (c) redesigned accumulator proven inside $\pi$.
Consequence if unresolved: Cannot claim count/order hiding in v1.
Safe fallback exists: yes, v1 treats $N$/order as visible.

Item: Hidden commitment list $\{C_i\}$ mode
Why it matters: If commitments are not provided, the verifier cannot recompute $R$ and $\alpha_i$ unless a proof of the accumulator relation is included.
Exact decision needed: Whether to introduce an accumulator-in-proof mode, and if so, which relation and proof system.
Allowed options: (a) no hidden $\{C_i\}$ (v1); (b) prove hash-chain relation inside $\pi$; (c) different accumulator design.
Consequence if unresolved: hidden-$\{C_i\}$ remains unsupported.
Safe fallback exists: yes, embedded list in v1.

Item: Refresh/rebase/compaction contract semantics
Why it matters: revocation and lifecycle safety depend on a precise state machine.
Exact decision needed: define contract state variables, events, and query interface.
Allowed options: TBD (separate contract spec).
Consequence if unresolved: lifecycle security properties cannot be stated as settled.
Safe fallback exists: yes, spec does not claim forward-security properties.

Item: Non-linear proof profile
Why it matters: real-world policies often involve non-linear constraints.
Exact decision needed: select and fully specify a proof system and compilation format (for example Bulletproofs circuits), including transcript schedule and statement binding.
Allowed options: keep v1 linear-only; adopt Bulletproofs-style or related discrete-log arithmetic-circuit proofs with complete mapping; adopt a folding-based IVC scheme such as Nova or HyperNova with explicit reductions. [6], [15], [16], [17]
Consequence if unresolved: non-linear support remains non-normative.
Safe fallback exists: yes, v1 linear-only.

## Revision Change Log

Narrative alignment:

- Tightened narrative claims to match explicit leakage and verifier obligations.
- Clarified pk semantics as application-scoped and transcript-bound.

Specification-level changes:

- Deterministic CBOR profile, strict reject rules, and wire schemas are now normative, resolving transcript-binding ambiguity. [1]
- Ciphersuite, suite identifiers, H2G/H2S, and DST registry are frozen to RFC 9380/9496 and SHA-512. [2], [3], [4]
- Commitment Profile V2 is made dimensionally consistent.
- Proof system fully specified as $\pi_{\mathrm{link}} + \pi_{\mathrm{cons}}$; removed infeasible IPA-extraction language and replaced with proved relations grounded in Schnorr NIZK. [5]
- Added computed $N = 8$ test vectors (linear end-to-end plus non-linear folding example).

Claims downgraded to goals:

- Count/order hiding by default.
- Hidden commitment list mode.
- Any lifecycle security depending on refresh/rebase/compaction.

Items moved to the open-decision set:

- Contract semantics for refresh/rebase/compaction.
- Full non-linear proof profile.

## References

1. C. Bormann and P. Hoffman, “Concise Binary Object Representation (CBOR),” RFC 8949, December 2020. <https://www.rfc-editor.org/rfc/rfc8949.html>
2. A. Faz-Hernandez, S. Scott, N. Sullivan, R. S. Wahby, and C. A. Wood, “Hashing to Elliptic Curves,” RFC 9380, August 2023. <https://www.rfc-editor.org/rfc/rfc9380.html>
3. H. de Valence, J. Grigg, G. Tankersley, F. Valsorda, and I. Lovecruft, “The ristretto255 and decaf448 Groups,” RFC 9496, December 2023. <https://www.rfc-editor.org/rfc/rfc9496.html>
4. National Institute of Standards and Technology, “Secure Hash Standard,” FIPS 180-4, August 2015. <https://www.nist.gov/publications/secure-hash-standard>
5. M. Nir, “Schnorr Non-interactive Zero-Knowledge Proof,” RFC 8235, September 2017. <https://datatracker.ietf.org/doc/html/rfc8235>
6. B. Bunz, J. Bootle, D. Boneh, A. Poelstra, P. Wuille, and G. Maxwell, “Bulletproofs: Short Proofs for Confidential Transactions and More,” Cryptology ePrint Archive, Paper 2017/1066. <https://eprint.iacr.org/2017/1066>
7. NESSA qFold-EC Revised Whitepaper Package, archived local PDF authority file `Archive/Superseded Whitepapers and Audits/2026-05-05/NESSA qFold-EC Revised Whitepaper.pdf`, SHA-256 `e5d1e8278ce92f13321adb89e0ec1e14d9ae301935a76406b8c3a9fab7cec048`.
8. Audit-backed whitepaper update file `whitepaper-230326-audit.md`, local audit authority file, SHA-256 `4dc05fdbf190567bb41ba6af296da75104fb527f67599801edd5645a7cb056fc`.
9. T. P. Pedersen, “Non-Interactive and Information-Theoretic Secure Verifiable Secret Sharing,” CRYPTO 1991, LNCS 576, pp. 129-140. <https://iacr.org/cryptodb/data/paper.php?pubkey=1671>
10. Standards for Efficient Cryptography Group, “SEC 1: Elliptic Curve Cryptography,” Version 2.0, May 2009. <https://www.secg.org/sec1-v2.pdf>
11. Standards for Efficient Cryptography Group, “SEC 2: Recommended Elliptic Curve Domain Parameters,” Version 2.0, January 2010. <https://www.secg.org/sec2-v2.pdf>
12. National Institute of Standards and Technology, “Recommendations for Discrete Logarithm-Based Cryptography: Elliptic Curve Domain Parameters,” SP 800-186, February 2023. <https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-186.pdf>
13. A. Rundgren, B. Jordan, and S. Erdtman, “JSON Canonicalization Scheme (JCS),” RFC 8785, June 2020. <https://www.rfc-editor.org/rfc/rfc8785.html>
14. dalek-cryptography, “Merlin: Composable Proof Transcripts for Public-Coin Arguments of Knowledge.” <https://github.com/dalek-cryptography/merlin>
15. A. Kothapalli, S. Setty, and I. Tzialla, “Nova: Recursive Zero-Knowledge Arguments from Folding Schemes,” Cryptology ePrint Archive, Paper 2021/370. <https://eprint.iacr.org/2021/370>
16. A. Kothapalli and S. Setty, “HyperNova: Recursive Arguments for Customizable Constraint Systems,” Cryptology ePrint Archive, Paper 2023/573. <https://eprint.iacr.org/2023/573>
17. J. Bootle, A. Cerulli, P. Chaidos, J. Groth, and C. Petit, “Efficient Zero-Knowledge Arguments for Arithmetic Circuits in the Discrete Log Setting,” Cryptology ePrint Archive, Paper 2016/263. <https://eprint.iacr.org/2016/263>

# NESSA qFold-EC Core Protocol Formulas

Reference-quality formula sheet for the NESSA qFold-EC v1 proof system.
The immediate source of truth for these formulas is
`nessa_qfold.py`.

---

## 1. Constants And Groups

| Symbol | Value | Source |
|---|---|---|
| $G$ | Ristretto255 base point, RFC 9496 | `point_base_mul(SCALAR_ONE)` |
| $L$ | $2^{252} + 27742317777372353535851937790883648493$ | `L` |
| $\widetilde{0}$ | 32-byte all-zero identity point | `IDENTITY` |
| $\mathrm{H2G}$ | `hash_to_ristretto255` via `expand_message_xmd(SHA-512)`, RFC 9380 | `h2g()` |
| $\mathrm{H2S}$ | `hash_to_field` over $\mathbb{F}_{L}$ via `expand_message_xmd(SHA-512)` | `h2s()` |
| $\mathrm{EncCBOR}$ | Deterministic CBOR encoding, RFC 8949 canonical map ordering | `cbor_encode()` |

---

## 2. Generator Derivation

For message dimension $d$, qFold-EC derives two vector generator families and
two policy generators. For each $0 \le j < d$:

$$
B^{m}_{j} = \mathrm{H2G}\left(\mathsf{DST\\_BASE\\_BM}, \mathrm{EncCBOR}\left([\text{base}, \text{Bm}, j + 1, \mathsf{PROTOCOL\\_VERSION}]\right)\right)
$$

$$
B^{m}_{j} = \mathrm{H2G}\left(\mathsf{DST\\_BASE\\_BM}, \mathrm{EncCBOR}\left([\text{base}, \text{Bm}, j + 1, \mathsf{PROTOCOL\\_VERSION}]\right)\right)
$$

The policy generators are:

$$
G_{\mathrm{pol}} = \mathrm{H2G}\left(\mathsf{DST\\_BASE\\_GPOL}, \mathrm{EncCBOR}\left([\text{base}, \text{Gpol}, 0, \mathsf{PROTOCOL\\_VERSION}]\right)\right)
$$

$$
H_{\mathrm{pol}} = \mathrm{H2G}\left(\mathsf{DST\\_BASE\\_HPOL}, \mathrm{EncCBOR}\left([\text{base}, \text{Hpol}, 0, \mathsf{PROTOCOL\\_VERSION}]\right)\right)
$$

If a derived generator is the identity or duplicates an existing generator, the
implementation retries with a counter appended to the encoded preimage. The
accepted set must contain $2d + 2$ distinct non-identity points.

- Source: `derive_generator()`, `derive_generators()`, `validate_generator_set()`

---

## 3. Commitment Profile V2

Each event row has scalar-field coordinates:

$$
m_{i} = (m_{i,0}, \ldots, m_{i,d-1})
$$

and blinding vector:

$$
\rho_{i} = (\rho_{i,0}, \ldots, \rho_{i,d-1})
$$

The row commitment is:

$$
C_{i} = \mathrm{Com}_{\mathrm{V2}}(m_{i}; \rho_{i})
$$

$$
C_{i} = \sum_{j=0}^{d-1}\left(\rho_{i,j}B^{r}_{j} + m_{i,j}B^{m}_{j}\right)
$$

The commitment is additively homomorphic:

$$
\mathrm{Com}_{\mathrm{V2}}(a; r) + \mathrm{Com}_{\mathrm{V2}}(b; s) = \mathrm{Com}_{\mathrm{V2}}(a + b; r + s)
$$

- Source: `commit_v2()`

---

## 4. Transcript Root Chain

Tags and commitments are bound into a deterministic hash chain.

| Tag key | Value |
|---|---|
| $\mathsf{tags}[0]$ | $\mathsf{PROTOCOL\\_VERSION\\_NUMBER}$ |
| $\mathsf{tags}[1]$ | $\mathsf{PROTOCOL\_VERSION}$ |
| $\mathsf{tags}[2]$ | $\mathsf{RFC9380\_H2G\_ID}$ |
| $\mathsf{tags}[3]$ | $\mathsf{encoding\_id}$ |
| $\mathsf{tags}[4]$ | $\mathsf{encoding\_hash}$ |
| $\mathsf{tags}[5]$ | $d$ |
| $\mathsf{tags}[6]$ | $\mathsf{policy\_id}$ |
| $\mathsf{tags}[7]$ | $\mathsf{policy\_hash}$ |
| $\mathsf{tags}[8]$ | $\mathsf{k\_rows}$ |

If present, `transcript_seed` is added as tag key `9`.

$$
\mathsf{tags\_hash} = \mathrm{SHA512}\left(\mathrm{EncCBOR}(\mathsf{tags})\right)
$$

$$
R_{0} = \mathrm{SHA512}\left(\mathrm{EncCBOR}\left([\text{NESSA-EC:v1:R0}, \mathsf{tags\_hash}]\right)\right)
$$

For $1 \le i \le N$:

$$
R_{i} = \mathrm{SHA512}\left(\mathrm{EncCBOR}\left([\text{NESSA-EC:v1:Ri}, i, R_{i-1}, C_{i}]\right)\right)
$$

- Source: `build_tags()`, `build_transcript()`

---

## 5. Fold Weights

Fold weights are Fiat-Shamir scalar challenges derived from the final
transcript root $R_{N}$.

For $1 \le i \le N$:

$$
\alpha_{i} = \mathrm{H2S}\left(\mathsf{DST\_ALPHA}, \mathrm{EncCBOR}\left([\text{alpha}, R_{N}, i]\right)\right)
$$

- Source: `compute_alpha()`

---

## 6. Commitment Folding

The proof folds $N$ row commitments into one commitment:

$$
C^{\star} = \sum_{i=1}^{N}\alpha_{i}C_{i}
$$

The witness and blinding vectors fold coordinate-wise. For $0 \le j < d$:

$$
(m^{\star})_{j} = \sum_{i=1}^{N}\alpha_{i}m_{i,j} \pmod{L}
$$

$$
(\rho^{\star})_{j} = \sum_{i=1}^{N}\alpha_{i}\rho_{i,j} \pmod{L}
$$

The folded commitment must satisfy:

$$
C^{\star} = \mathrm{Com}_{\mathrm{V2}}(m^{\star}; \rho^{\star})
$$

- Source: `fold_commitments()`, `fold_witnesses()`, `fold_randomness()`

---

## 7. Policy Commitments

For each folded coordinate, the prover creates a policy commitment. For
$0 \le j < d$:

$$
V_{j} = \gamma_{j}G_{\mathrm{pol}} + (m^{\star})_{j}H_{\mathrm{pol}}
$$

Here $\gamma_{j}$ is a fresh blinding scalar.

- Source: `policy_commit()`

---

## 8. Proof Context

The proof context binds the proof to application-level semantics:

$$
\mathsf{proof\\_context} = \mathrm{EncCBOR}\left([\text{proof\\_context}, \mathsf{tags\\_hash}, R_{N}, N, d, \mathsf{context\\_label}]\right)
$$

`context_label` is normalized by `normalize_context_binding()` before encoding.

- Source: `build_proof_context()`

---

## 9. `pi_link` - Linkage Proof

Statement: the prover knows $(m^{\star}, \rho^{\star}, \gamma)$ such that:

$$
C^{\star} = \mathrm{Com}_{\mathrm{V2}}(m^{\star}; \rho^{\star})
$$

For $0 \le j < d$:

$$
V_{j} = \gamma_{j}G_{\mathrm{pol}} + (m^{\star})_{j}H_{\mathrm{pol}}
$$

Nonce commitments:

$$
T_{\mathrm{commit}} = \mathrm{Com}_{\mathrm{V2}}(k_{m}; k_{\rho})
$$

For $0 \le j < d$:

$$
T_{j} = k_{\gamma,j}G_{\mathrm{pol}} + k_{m,j}H_{\mathrm{pol}}
$$

Challenge:

$$
c = \mathrm{H2S}\left(\mathsf{DST\_LINK}, \mathrm{EncCBOR}\left([\text{link}, \mathsf{tags\_hash}, R_{N}, C^{\star}, \mathsf{V\_list}, T_{\mathrm{commit}}, \mathsf{T\_policy}]\right)\right)
$$

Responses, all modulo $L$:

$$
z_{m,j} = k_{m,j} + c(m^{\star})_{j} \pmod{L}
$$

$$
z_{\rho,j} = k_{\rho,j} + c(\rho^{\star})_{j} \pmod{L}
$$

$$
z_{\gamma,j} = k_{\gamma,j} + c\gamma_{j} \pmod{L}
$$

Verification recomputes $c$ and checks:

$$
\mathrm{Com}_{\mathrm{V2}}(z_{m}; z_{\rho}) = T_{\mathrm{commit}} + cC^{\star}
$$

For $0 \le j < d$:

$$
z_{\gamma,j}G_{\mathrm{pol}} + z_{m,j}H_{\mathrm{pol}} = T_{j} + cV_{j}
$$

- Source: `prove_link()`, `verify_link()`
- Proof component size: $(d + 1)$ points, $3d$ scalars, and 1 challenge scalar.

---

## 10. `pi_cons_linear` - Linear Constraint Proof

Linear policies are compiled and transcript-compressed before proving. For
compressed coefficients $a_{j}$ and compressed target $t$, the prover shows:

$$
\sum_{j=0}^{d-1}a_{j}(m^{\star})_{j} = t \pmod{L}
$$

Define:

$$
\gamma_{\mathrm{res}} = \sum_{j=0}^{d-1}a_{j}\gamma_{j} \pmod{L}
$$

$$
W = \sum_{j=0}^{d-1}a_{j}V_{j} - tH_{\mathrm{pol}}
$$

When the linear relation holds:

$$
W = \gamma_{\mathrm{res}}G_{\mathrm{pol}}
$$

The prover sends:

$$
T = kG_{\mathrm{pol}}
$$

$$
c = \mathrm{H2S}\left(\mathsf{DST\_CONS}, \mathrm{EncCBOR}\left([\text{cons}, \mathsf{tags\_hash}, R_{N}, \mathsf{policy\_hash}, W, T]\right)\right)
$$

$$
z = k + c\gamma_{\mathrm{res}} \pmod{L}
$$

Verification recomputes $c$ and checks:

$$
zG_{\mathrm{pol}} = T + cW
$$

- Source: `compressed_linear_terms()`, `linear_constraint_W()`, `prove_cons_linear()`, `verify_cons_linear()`
- Proof component size: 1 point, 1 response scalar, and 1 challenge scalar.

---

## 11. `pi_cons_nonlinear` - Multiplicative Constraint Proof

The EC non-linear proof object exists in the Python reference implementation,
but EC v1 remains production-specified for linear policies only. Treat this
section as implementation-aligned reference/demo math, not a production EC v1
claim.

Given folded committed values $(L^{\star}, R^{\star}, O^{\star})$ in
$(V_{L}, V_{R}, V_{O})$, prove:

$$
L^{\star}R^{\star} = O^{\star} + E^{\star} \pmod{L}
$$

The cross-term error is accumulated by `nonlinear_fold()`:

$$
w_{0} = \mathrm{int}(\alpha_{0})
$$

$$
L_{\mathrm{acc}} = w_{0}L_{0}
$$

$$
R_{\mathrm{acc}} = w_{0}R_{0}
$$

$$
O_{\mathrm{acc}} = w_{0}O_{0}
$$

$$
E_{\mathrm{acc}} = (w_{0}^{2} - w_{0})O_{0}
$$

For each row $1 \le i < N$:

$$
w_{i} = \mathrm{int}(\alpha_{i})
$$

$$
T_{i} = L_{\mathrm{acc}}R_{i} + L_{i}R_{\mathrm{acc}}
$$

$$
E_{\mathrm{acc}} \leftarrow E_{\mathrm{acc}} + w_{i}T_{i} + (w_{i}^{2} - w_{i})O_{i}
$$

$$
L_{\mathrm{acc}} \leftarrow L_{\mathrm{acc}} + w_{i}L_{i}
$$

$$
R_{\mathrm{acc}} \leftarrow R_{\mathrm{acc}} + w_{i}R_{i}
$$

$$
O_{\mathrm{acc}} \leftarrow O_{\mathrm{acc}} + w_{i}O_{i}
$$

All recurrence operations are modulo $L$, and
$E^{\star} = E_{\mathrm{acc}}$ at termination.

The prover commits to the error:

$$
C_{E} = r_{E}G_{\mathrm{pol}} + E^{\star}H_{\mathrm{pol}}
$$

Nonce commitments:

$$
T_{L} = k_{\gamma,L}G_{\mathrm{pol}} + k_{L}H_{\mathrm{pol}}
$$

$$
T_{R} = k_{\gamma,R}G_{\mathrm{pol}} + k_{R}H_{\mathrm{pol}}
$$

$$
T_{O} = k_{\gamma,O}G_{\mathrm{pol}} + k_{O}H_{\mathrm{pol}}
$$

$$
T_{E} = k_{rE}G_{\mathrm{pol}} + k_{E}H_{\mathrm{pol}}
$$

$$
T_{\mathrm{mulBase}} = k_{\mathrm{mulBase}}G_{\mathrm{pol}} + (k_{L}k_{R})H_{\mathrm{pol}}
$$

$$
T_{\mathrm{mulCross}} = k_{\mathrm{mulCross}}G_{\mathrm{pol}} + (k_{L}R^{\star} + k_{R}L^{\star})H_{\mathrm{pol}}
$$

Challenge:

$$
c = \mathrm{H2S}\left(\mathsf{DST\_CONS}, \mathrm{EncCBOR}\left([\text{NESSA-EC:v1:schnorr}, \mathsf{proof\_context}, \mathsf{all\_T}, \mathsf{all\_P}]\right)\right)
$$

where:

$$
\mathsf{all\_T} = [T_{L}, T_{R}, T_{O}, T_{E}, T_{\mathrm{mulBase}}, T_{\mathrm{mulCross}}]
$$

$$
\mathsf{all\_P} = [V_{L}, V_{R}, V_{O}, C_{E}]
$$

Responses:

$$
z_{L} = k_{L} + cL^{\star}
$$

$$
z_{R} = k_{R} + cR^{\star}
$$

$$
z_{O} = k_{O} + cO^{\star}
$$

$$
z_{E} = k_{E} + cE^{\star}
$$

$$
z_{\gamma,L} = k_{\gamma,L} + c\gamma_{L}
$$

$$
z_{\gamma,R} = k_{\gamma,R} + c\gamma_{R}
$$

$$
z_{\gamma,O} = k_{\gamma,O} + c\gamma_{O}
$$

$$
z_{rE} = k_{rE} + cr_{E}
$$

$$
z_{\mathrm{mulBlind}} = k_{\mathrm{mulBase}} + ck_{\mathrm{mulCross}} + c^{2}(\gamma_{O} + r_{E})
$$

Verification recomputes $c$ and checks:

$$
z_{\gamma,L}G_{\mathrm{pol}} + z_{L}H_{\mathrm{pol}} = T_{L} + cV_{L}
$$

$$
z_{\gamma,R}G_{\mathrm{pol}} + z_{R}H_{\mathrm{pol}} = T_{R} + cV_{R}
$$

$$
z_{\gamma,O}G_{\mathrm{pol}} + z_{O}H_{\mathrm{pol}} = T_{O} + cV_{O}
$$

$$
z_{rE}G_{\mathrm{pol}} + z_{E}H_{\mathrm{pol}} = T_{E} + cC_{E}
$$

$$
z_{\mathrm{mulBlind}}G_{\mathrm{pol}} + (z_{L}z_{R})H_{\mathrm{pol}} = T_{\mathrm{mulBase}} + cT_{\mathrm{mulCross}} + c^{2}(V_{O} + C_{E})
$$

- Source: `nonlinear_fold()`, `prove_cons_nonlinear()`, `verify_cons_nonlinear()`
- Proof component size: 7 points, 9 response scalars, and 1 challenge scalar.

---

## 12. Complete Proof Object

The Python dataclass `NessaProof` has public data and proof components:

$$
\pi = (\pi_{\mathrm{link}}, \pi_{\mathrm{cons}})
$$

The linkage proof is:

$$
\pi_{\mathrm{link}} = (T_{\mathrm{commit}}, \mathsf{T\_policy}, z_{m}, z_{\rho}, z_{\gamma}, c)
$$

The linear constraint proof is:

$$
\pi_{\mathrm{consLinear}} = (T, z, c)
$$

The non-linear constraint proof is:

$$
\pi_{\mathrm{consNonlinear}} = (C_{E}, T_{L}, T_{R}, T_{O}, T_{E}, T_{\mathrm{mulBase}}, T_{\mathrm{mulCross}}, z_{L}, z_{R}, z_{O}, z_{E}, z_{\gamma,L}, z_{\gamma,R}, z_{\gamma,O}, z_{rE}, z_{\mathrm{mulBlind}}, c)
$$

For a linear proof with dimension $d$, `NessaProof.byte_size()` counts:

$$
\mathrm{publicPoints} = d + 1
$$

$$
\pi_{\mathrm{link}} = (d + 1)\text{ points} + 3d\text{ scalars} + 1\text{ scalar}
$$

$$
\pi_{\mathrm{consLinear}} = 1\text{ point} + 2\text{ scalars}
$$

Since points and scalars are 32 bytes:

$$
\text{linear proof bytes} = (5d + 6) \cdot 32
$$

For $d = 9$:

$$
(5 \cdot 9 + 6) \cdot 32 = 51 \cdot 32 = 1632\text{ bytes}
$$

For a non-linear proof, `NessaProof.byte_size()` adds the non-linear component:

$$
\text{nonlinear proof bytes} = (5d + 20) \cdot 32
$$

For $d = 3$:

$$
(5 \cdot 3 + 20) \cdot 32 = 35 \cdot 32 = 1120\text{ bytes}
$$

- Source: `NessaProof.byte_size()`

---

## 13. Security Assumptions

| Assumption | Description |
|---|---|
| Discrete logarithm | Discrete logarithm in Ristretto255 is hard. |
| Generator independence | The $2d + 2$ accepted generators are independently derived with no known discrete-log relations. |
| Fiat-Shamir random oracle | SHA-512, H2G, and H2S behave as random oracles for transcript-derived challenges. |
| CBOR canonicality | Deterministic CBOR encoding gives prover and verifier the same transcript preimages. |
| Fresh prover randomness | Live proving requires fresh unpredictable prover randomness; deterministic seeds are for fixed vectors and tests only. |
| Application domain validation | Application coordinate domains must be validated before scalar-field reduction to avoid wraparound and overflow mistakes. |

- Source: `SECURITY_ASSUMPTIONS`

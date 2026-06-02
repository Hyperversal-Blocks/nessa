# NESSA
NESSA is a privacy-first, local-first protocol layer designed for identity, access, and credential management. 

It enables a user or device to compress numerous sensitive security events, such as enrollments, key rotations, and policy checks, into a single verifiable object. A verifier can inspect this object to confirm policy compliance without accessing the underlying event data or leaking metadata. 

This capability is powered by **qFold**, a folding transform that aggregates event commitments and generates a succinct proof of correctness. NESSA maintains a consistent application experience across two cryptographic backends. 

**qFold-EC** is an elliptic-curve track available for immediate deployment, while **qFold-PQ** provides a post-quantum migration path. This dual-engine design ensures that the user experience, APIs, and privacy semantics remain stable even as the underlying cryptography evolves.

## One control layer for users, verifiers, apps, agents, wallets, and chains.
Modern verification creates privacy debt.

Apps accumulate cryptographic breadcrumbs:
- identity records,
- credentials,
- tokens,
- logs,
- signatures,
- user context.

Verifying them later is brittle, leaky, and replay-prone.

Stronger verification often becomes more data collection.

## Prove, Don’t Show
Using a proof-controller, users can share the proof of permission without disclosing the whole identity behind it without exposing raw credentials.

**NESSA is a self-hosted Proof Controller.**

The verifier asks for proof.
The user approves what can be shown.
The verifier receives a proof-bound response, not the underlying data.

Optional on-chain anchoring and enforcement can be added where needed.

## qFold makes private proof state compact and verifiable.
qFold is the proof infrastructure under NESSA.
It folds many private committed events into one proof package:
- credentials,
- permissions,
- delegations,
- session facts,
- state change → folded proof state → verifiable proof package

The verifier checks the proof result and required metadata without receiving the private event history.

## NESSA as a Proof-Controller
| Existing Approach | Usually Gives | NESSA Adds |
|---|---|---|
| OAuth / IAM | Access tokens | Proof-based responses |
| Credential wallets | Credential presentation | Request control and receipts |
| On-chain attestations | Public verification | Private proof with optional public anchoring |
| ZK apps | Custom proof logic | Reusable proof controller |
| Agent permissions | Signed delegation | Context-bound proof authority |

**NESSA is a proof substrate, Proof Controller, and then some.**
- Proof Controller for users.
- Proof infrastructure for verifiers.
- Adapter layer for apps, wallets, agents, and chains.
- Privacy layer for controlled disclosure, receipts, and optional anchoring.

**NESSA controls what you prove.**
It sits between users, apps, agents, wallets, and chains so verifiers can receive proof without collecting the underlying private data.

## Proof-Control Layer
NESSA combines four things that are usually separate:
- User-controlled proof requests
- Context and transcript-bound proof generation
- Folding-based proof compression using qFold
- Verifier-ready proof responses

### How proving works
- App asks for proof
- NESSA shows the user what is being requested
- User approves, denies, limits or delegates
- qFold generates the proof package
- Verifier checks the proof response
- User receives a privacy receipt

### How verification works
- The proof is valid
- The request context matches
- The policy matches
- The proof is well-formed and fresh
- The proof has not been reused
- The response came from expected controller path

**The verifier will not receive:**
- Raw credentials
- Full identity profile
- Private event history
- Unrelated permissions
- Internal user context

## Sustainability
NESSA is open-source infrastructure. The project may be sustained through grants, hosted verifier services, enterprise integrations, support contracts, and ecosystem partnerships.

The core proof request, verifier response, and proof controller specifications are intended to remain open and self-hostable.

## Unified Surface for Dual Cryptographic Engines
NESSA provides a consistent API and semantic surface that is independent of the underlying cryptographic engine. **qFold-EC** uses Pedersen-style commitments and inner-product arguments on elliptic curves. **qFold-PQ** uses Module-LWE commitments and folded lattice sigma-protocols for post-quantum resilience. This ensures that migration is a managed process, not a complete product rewrite. Nessa also provides ecosystem adapters for prover semantics.

## Positioning

NESSA is a proof controller for privacy-preserving identity, credential verification, access control, and selective disclosure.

It sits between verifiers, users, wallets, applications, and proof backends, allowing verifiers to request proofs without receiving unnecessary personal data.

NESSA is designed for developers building:
- privacy-preserving login flows
- credential-based access systems
- verifier-controlled proof requests
- reusable ZK authorization flows
- agent and delegation-based permission systems

## Current Progress

* NIC Karachi and Uniswap Hook Incubator validation
* `nessa.sh` website and demos are live
* Math and code behind linear folding and transcript binding completed with local tests
* [Whitepaper](https://github.com/Hyperversal-Blocks/nessa/blob/master/whitepaper.md) and v1 architecture completed
* Product and application diagrams completed
* UHI hook architecture designed
* Code-paper internal [audit](https://github.com/Hyperversal-Blocks/nessa/blob/master/whitepaper-230326-audit.md) and [profile-benchmarks](https://github.com/Hyperversal-Blocks/nessa/blob/master/benchmarks/qfold-ec/2026-05-07/README.md).

## Roadmap

* Proof Controller product demo
* Verifier SDKs, proof request templates, and integration guides
* qFold upgrade path for broader proof workflows
* qFold-PQ migration track
* App, wallet, agent, and chain integrations with optional on-chain anchors
* Adapter integrations and ecosystem expansion

## License
[GNU General Public License v3.0](https://github.com/Hyperversal-Blocks/nessa/blob/master/LICENSE)

## Documents
- `whitepaper.md` is the app-repo copy of the authoritative qFold-EC whitepaper
  maintained in the private Research authority directory.
- `whitepaper-230326-audit.md` is the companion audit-backed update report.
- `benchmarks/qfold-ec/2026-05-07/` is the app-repo copy of the qFold-EC
  benchmark and generated verification bundle.
- `archive/whitepaper-previous-2026-05-07.md` preserves the previous app-level
  whitepaper that this copy replaced.
- [NESSA website and demos](https://nessa.sh/v1/)
- [Letter of Association with National Incubation Center](https://drive.google.com/file/d/1iDUh_AQjsN8puyP8zXosPuUcbNIHPS_T/view?usp=drive_link)

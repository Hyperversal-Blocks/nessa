# NESSA qFold-EC Benchmark Bundle - 2026-05-07

This directory is the `apps/nessa` repository copy of the qFold-EC generated
benchmark and verification bundle. The authoritative copy is maintained in the
private Research authority directory.

The source implementation for these artifacts is the vendored Python qFold-EC
implementation used to generate this snapshot.

## Contents

- `manifest/docs_bundle_manifest.json`: generated bundle manifest.
- `protocol/test_vectors/test_vectors_output.json`: qFold protocol test vectors.
- `protocol/verification/verification_report.json`: machine-readable qFold verification report.
- `protocol/verification/verification_report.txt`: human-readable qFold verification report.
- `asc_ad_demo/audit/asc_ad_artifact_manifest.json`: ASC artifact manifest.
- `asc_ad_demo/audit/asc_ad_benchmark_rows.json`: qFold-backed ASC benchmark rows.
- `asc_ad_demo/audit/asc_ad_benchmark_validity.json`: benchmark validity checks.
- `asc_ad_demo/reports/asc_ad_benchmark_report.json`: aggregate benchmark report.
- `asc_ad_demo/reports/asc_ad_benchmark_report.txt`: human-readable benchmark report.

## Snapshot

- Deterministic bundle: yes.
- Benchmark enabled: yes.
- Protocol verification report: 80 checks, 80 passed, 0 failed.
- Protocol vectors: `TV-LIN-8` proof size 832 bytes; `TV-R1CS-8` proof size 1120 bytes.
- ASC benchmark rows: 13.
- ASC benchmark proof size: constant 1632 bytes in this profile.

See `SHA256SUMS.txt` for artifact hashes.

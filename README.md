# paramind

`paramind` is a passive CLI for bug bounty reconnaissance. It reads URLs from `stdin`, extracts query parameters, classifies them into offensive semantic categories, and highlights which URLs deserve manual testing first.

It does not send HTTP requests, does not fuzz, and does not attempt exploitation.

## What It Does

- Reads one URL per line from `stdin`
- Accepts only valid `http` and `https` URLs
- Skips empty lines, invalid URLs, static asset URLs, and URLs without query strings
- Supports SPA hash routes (`#/path?...` and `#!/path?...`) when no backend query is present
- Deduplicates URLs (exact by default, or value-type-aware signature via `-dedupe`)
- Classifies parameters by semantic class
- Assigns a confidence level based on exact, normalized, or partial matches
- Emits results immediately as it processes input
- Prints execution stats at the end (including a `Duplicates` counter)

The processing model is streaming-based and memory-conscious: URLs are processed one at a time, without accumulating the full input set in memory.

## Semantic Classes

Classes are prioritized in this order:

1. `auth`
2. `redirect`
3. `ssrf`
4. `file`
5. `id`
6. `sqli`
7. `xss`
8. `debug`

Each classified parameter includes:

- parameter name
- value
- class
- confidence
- suggested vulnerability hypotheses to test manually

The classifier includes common English keys plus localized aliases useful for BR/PT and some ES-style parameter names such as `ordem`, `retorno_url`, `arquivo`, `sessão`, `usuario`, `destino`, and related variants.

## Installation

Build locally:

```bash
go build -o paramind ./cmd/paramind
```

Run from the repository root:

```bash
./paramind
```

Or install somewhere in your `PATH`:

```bash
go build -o /usr/local/bin/paramind ./cmd/paramind
```

## Basic Usage

Feed URLs through `stdin`:

```bash
cat urls.txt | ./paramind
```

With passive recon tools:

```bash
waybackurls example.com | ./paramind
gau example.com | ./paramind
```

Combined workflow:

```bash
(waybackurls example.com; gau example.com) | sort -u | ./paramind
```

## Example

Input:

```bash
printf '%s\n' \
  'https://app.example.com/login?next=%2Fdashboard&id=12&token=abc123' \
  'https://app.example.com/report?retorno_url=%2Fpainel&arquivo=relatorio.pdf&debug=true' \
  | ./paramind
```

Output:

```text
URL: https://app.example.com/login?next=%2Fdashboard&id=12&token=abc123

  Param: next
  Value: /dashboard
  Class: redirect
  Confidence: high
  Test: open_redirect, redirect_validation_bypass, oauth_redirect_abuse, token_leak_via_redirect

  Param: id
  Value: 12
  Class: id
  Confidence: high
  Test: idor, enumeration, broken_access_control

  Param: token
  Value: abc123
  Class: auth
  Confidence: high
  Test: token_leakage, session_fixation, account_takeover_vector, weak_auth_flow
```

## Flags

```text
-json
    Emit JSON Lines output

-all
    Include unclassified parameters

-min-confidence low|medium|high
    Only show findings at or above the given confidence

-silent
    Only print matching URLs

-category auth,redirect,ssrf,file,id,sqli,xss,debug
    Restrict output to one or more classes

-dedupe exact|signature
    How to deduplicate URLs. Default is `exact` (full-URL dedup), so
    `?id=1` and `?id=2` both appear.
    `signature` dedupes on host + path + sorted param keys annotated with a
    per-value type bucket: `n` numeric, `s` string, `e` empty. Sequential
    noise like `?id=1, ?id=2, ?id=3` collapses to one, but type variation
    like `?id=admin` or `?id=` is still preserved as a separate URL.
```

## JSON Output

Use `-json` to emit JSONL / NDJSON, one object per line:

```bash
(waybackurls example.com; gau example.com) | ./paramind -json
```

Example line:

```json
{"url":"https://example.com/login?next=%2Fdashboard&id=12","findings":[{"param":"next","value":"/dashboard","class":"redirect","confidence":"high","hypotheses":["open_redirect","redirect_validation_bypass","oauth_redirect_abuse","token_leak_via_redirect"]},{"param":"id","value":"12","class":"id","confidence":"high","hypotheses":["idor","enumeration","broken_access_control"]}]}
```

Execution stats are written to `stderr`, not `stdout`, so they do not corrupt JSON output. Save them separately if needed:

```bash
(waybackurls example.com; gau example.com) | ./paramind -json 1> findings.jsonl 2> stats.txt
```

Inspect results:

```bash
cat findings.jsonl
jq . findings.jsonl
```

## Useful Workflows

High-signal classes only:

```bash
(waybackurls example.com; gau example.com) | ./paramind -category auth,redirect,ssrf,file,id -min-confidence medium
```

Show only URLs that matched:

```bash
(waybackurls example.com; gau example.com) | ./paramind -silent
```

Include unclassified parameters for triage:

```bash
(waybackurls example.com; gau example.com) | ./paramind -all
```

Save readable output plus stats:

```bash
(waybackurls example.com; gau example.com) | ./paramind > findings.txt 2> stats.txt
```

## Processing Rules

For each line, `paramind`:

1. trims whitespace
2. skips empty lines
3. parses the URL with `net/url`
4. ignores invalid URLs
5. accepts only `http` and `https`
6. skips URLs without query parameters
7. filters common static file extensions
8. extracts query parameters
9. deduplicates URLs
10. classifies parameters
11. applies confidence and category filters
12. emits output immediately

Confidence rules:

- exact key match: `high`
- normalized match: `medium`
- partial match: `low`

Normalization lowercases keys and removes `_`, `-`, and common accents before comparison.

If multiple classes match a parameter, the higher confidence wins. So an exact match in a lower-priority class (e.g. `keyword` → `xss`) beats a partial match in a higher-priority class. Priority only breaks ties within the same confidence level.

When a URL has no backend query string, `paramind` looks at the fragment and accepts SPA hash routes that start with `/` or `!` and contain `?` (e.g. `https://app.example.com/#/admin/users?id=42`). Pure anchors like `#section` are still skipped. If both backend query and SPA fragment are present, the backend query wins.

## Scope

`paramind` is a semantic recon prioritizer, not a scanner.

It does not:

- make HTTP requests
- crawl
- fuzz
- verify vulnerabilities
- mutate payloads

It is intended to help you decide what to inspect manually next.

## Project Layout

```text
cmd/paramind/main.go
internal/parser/
internal/filter/
internal/classifier/
internal/output/
internal/model/
internal/dedupe/
```

# Review Go Code

Review the Go code changed or discussed in this conversation.

## What to Check

### Correctness
- Logic bugs, off-by-one errors, incorrect conditions
- Nil pointer dereferences (especially on map/slice/pointer access)
- Map writes on potentially nil maps (Go panics on write to nil map, not read)
- Integer overflow or unsafe type conversions
- Goroutine leaks or incorrect concurrency patterns (if applicable)

### Go Idioms
- Error handling: errors should be checked, not ignored with `_`
- Prefer early returns over deeply nested `if` blocks
- Redundant conditions (e.g., `if x == "" { return "" }` when `return x` would suffice)
- Unexported fields/functions that could/should be exported (or vice versa)
- Use of `fmt.Sprintf` where string concat is cleaner (and vice versa)

### Performance
- Unnecessary allocations inside hot loops
- Repeated slice/map lookups that could be cached in a variable
- Appending to a slice inside a loop that could pre-allocate with `make`

### Consistency with Codebase Patterns
- Does it follow the `wrapElement` / `DefineAccessorProperty` pattern used in `js/runtime.go`?
- Does error handling match the surrounding code style?
- Are new DOM properties placed in the correct tag-specific `if` block?

### Security
- No command injection, SQL injection, or XSS vectors introduced
- No user-controlled input passed unsanitized to shell or file operations

## Output Format

For each issue found:

**[Severity: Critical / Minor / Nit]** `file:line`
> Description of the issue and why it matters.
> Suggested fix (if applicable).

If no issues are found, say so clearly with a brief explanation of why the code looks correct.

## Scope

Focus only on code that was written, modified, or discussed in this conversation. Do not review unrelated code in the same files.

Run the following steps:

1. Run `git diff --cached --name-only` to see what files are staged.
2. Run `git diff --cached --stat` for a brief summary of changes.
3. Write a commit message of **at most 2 lines** based on what is staged:
   - Line 1: short imperative summary (max 72 chars), prefixed with `add:`, `fix:`, `update:`, or `remove:` depending on the nature of the change.
   - Line 2 (optional): one sentence of extra context only if truly needed.
4. Commit with that message using a HEREDOC.
5. Run `git push origin HEAD` to push to the current branch.
6. Confirm success with the commit hash and branch name.

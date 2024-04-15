
# sgit
## Intent
- `sgit` is a tool to synchronize version control across environments
### Usage
#### Prerequisites
```bash
export GITHUB_TOKEN=<github personal access token>
export GITHUB_ORG=<github organization>
export CODE_HOME_DIR=<path to root level code folder>
```
#### Running the command
```bash
sgit
```
#### What does this do?
1. ensures all organization repositories are cloned to `CODE_HOME_DIR` (organized in subfolders by language)
2. checks all local repositories for changes. If found, you'll be prompted to address them (`push`, `reset`, `stash` or `ignore`).
3. ensures all repositories and the current branch are up-to-date with the upstream


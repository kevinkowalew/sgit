# sgit 
## Intent
- `git` can get messy. `sgit` aims to bring order to the chaos.

# Usage
## Shim
To make `sgit` more pleasant to use, I point a `git` `alias` at the following shim.  This allows you to selectively run `sgit` commands alongside `git` without have to think about which binary to invoke (aka: `git ls`, `git clone`, `git delete`, `git init` all invoke `sgit` subcommands, while everything else leverages `git`)
```sh
if [[ $1 == "ls" || $1 == "clone" || $1 == "delete" || $1 == "init" ]]; then
	sgit "$@"
else
	git "$@"
fi
```

## Opinions
- `sgit` clones all repos to a specified `CODE_HOME_DIR` environment variable.
- As shown below, repos are organized in subdirectories by their respective languages
```
└── <CODE_HOME_DIR>/
    └── kevinkowalew/ 
        ├── go/
        │   └── my-go-server/
        │       ├── main.go
        │       ├── go.mod
        │       └── go.sum
        ├── java/
        │   └── my-java-server/
        │       ├── src/
        │       │   └── main/
        │       │       └── java/
        │       │           └── main.java
        │       └── pom.xml
        └── shell/
            └── dotfiles/
                ├── bash_profile.sh
                └── .tmuxconf
```

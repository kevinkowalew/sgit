# sgit 
## Intent
- `git` can get messy. `sgit` aims to bring order to the chaos.

# Usage
## Shim
To make `sgit` more pleasant to use, I point a `git` `alias` at the following shim.  This allows you to run `sgit` commands alongside traditional git without have to think about which binary to invoke.
```sh
if [[ $1 == "ls" || $1 == "sync" || $1 == "clone" || $1 == "delete" ]]; then
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

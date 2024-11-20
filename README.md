# sgit 
## Intent
- When working on `git` repos across multiple machines it can become hard to keep track of things.
- `sgit` aims to alleviate this issue, giving you visibility into the state of local `git` repositories against their remote counterparts.

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

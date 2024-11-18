# sgit 
## Intent
- When working on `git` repos across multiple machines it can become hard to keep track of things.
- `sgit` aims to alleviate this issue, giving you visibility into the state of local `git` repositories against their remote counterparts.

## Opinions
- `sgit` clones all repos to a specified `CODE_HOME_DIR` environment variable.
- As shown below, repos are organized in subdirectories by their respective languages
```
└── <CODE_HOME_DIR>/
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

## Subcommands
### `ls` - list local/remote repos
#### Output format
`<repo language>` `<full path to local repository>` `<state of repo>` `<is fork>`
```bash
$ sgit ls -langs java,go
java /Users/kevin/code/java/openapi-generator NotClonedLocally Fork
java /Users/kevin/code/java/http-client UpToDate
go /Users/kevin/code/go/sockets NoUpstreamRepo
go /Users/kevin/code/go/sgit HasUncommitedChanges
```
### `sync` - clones remote repos locally if needed
```bash
$ sgit sync -langs java,go
```

## Example Use Cases
- `sgit` embraces the unix philosophy. This enables more nuanced use cases to be achieved using piping.

```bash
# clone your repos to a new machine
sgit sync
```

```bash
# delete all local java forked repositories
sgit ls -langs java -forks true | awk `{print($2)}` | xargs rm -r
```

```bash
# find local go repos that you need to create remote repos for 
sgit ls -langs go -forks false -state NoRemoteRepo
```

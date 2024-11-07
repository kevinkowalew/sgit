# sgit 
## Intent
- When working with `git` repos on multiple machines it can become hard to keep track of the state of things in each respective environment.
- `sgit` aims to alleviate this issue, giving you visibility into the state of local `git` repositories against their remote counterparts.

## Directory Structure
- `sgit` clones all repos to a specified `CODE_HOME_DIR` environment variable.
- As shown below, repos are organized by name  
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
            └── -.tmuxconf
```

## Subcommands
### `ls` - outputs local/remote repos
#### Output format
`<repo language>` `<full path to local repository>` `<state of repo` `<shown if repo is a fork of another repo>`
```bash
$ sgit ls -langs java,go
java /Users/kevin/code/java/openapi-generator NotClonedLocally Fork
java /Users/kevin/code/java/http-client UpToDate
go /Users/kevin/code/go/sockets NoUpstreamRepo
go /Users/kevin/code/go/sgit HasUncommitedChanges
```
#### `sync` - clones remote repos locally if needed
```bash
$ sgit sync java,go
```

## Example Use Cases
- `sgit` embraces the unix philisophy. This enables more nuanced use cases to be achieved using piping.

```bash
# clone your repos to a new machine
sgit sync
```

```bash
# delete all local java forked repositories
sgit ls -langs java | grep Fork | awk `{print($2)}` | xargs rm -r
```

```bash
# find all local repos that you need to create upstream repos for 
sgit ls | grep NoUpstreamRepo
```

# API server for recmd-cli

## Introduction

This is the API server for `recmd-cli`. The list of handlers are:

- HandleDelete
- HandleAdd
- HandleSelect
- HandleSearch
- HandleRun
- HandleList

`recmd-dmn` must be started before `recmd-cli`. 

## Configuration

Two files will be created under `~/.recmd`. This is done automatically when `recmd-dmn` is started.

## Sample Output

```bash
$ ./recmd-dmn
2020/10/24 10:59:05 Starting server on :8999
```

### recmd_history.json

The list of commands in JSON format. If the file is not present, it will be created.

### recmd_secret

The file containing a secret. It is created every time `recmd-dmn` is started. The purpose is to provide a level of security as a "shared secret" between `recmd-dmn` and `recmd-cli`. 
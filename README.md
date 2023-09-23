# Backupper App <!-- omit from toc -->

Welcome to Backupper! The app for making backup easy and fast!

- [Features](#features)
- [Usage](#usage)
- [Watchers](#watchers)

## Features

* Create backup of files in the `"hot"` directory
* Select `watcher` of changes based on your expectations
* Copy is made by chunks, so large files are supported
* All non-regular files are ignore (e.g. dirs, device files)
* Files in backup dir supposed to be maintained only by app (**DO NOT CHANGE BACKUP DIR FILES POLICY =)**)
* Create file with name `delete_<filename>` and `<filename>` will be removed in `hot` and `backup` dirs

## Usage

```bash
# for build
> make build

# for run
> ./main

# for usage help
> ./main -h

# to create sample files check the Makefile
```

## Watchers

| Watcher | Description |
|-|-|
| syscal | Watcher that uses `syscall` with non-blocking IO. Fast, without "loss window", but could consume resources by triggering a lot of events |
| event | Watcher that uses `github.com/fsnotify/fsnotify`. Same `syscall` under the hood, produces more events than `syscall`, but more reliable |
| timer | Watcher that runs by ticker, uses much less resources, but has "loss window" (triggered once per sec) |

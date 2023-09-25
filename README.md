# Backupper App <!-- omit from toc -->

Welcome to Backupper! The app for making backup easy and fast!

- [Features](#features)
- [Basic usage](#basic-usage)
- [Backup mode](#backup-mode)
- [View logs mode](#view-logs-mode)
- [Testing](#testing)
- [Docker](#docker)
- [Watchers](#watchers)

## Features

* Create backup of files in the `"hot"` directory
* Select `watcher` of changes based on your expectations
* Copy is made by chunks, so large files are supported
* All non-regular files are ignored (e.g. dirs, device files)
* Files in backup dir supposed to be maintained only by app (**DO NOT CHANGE BACKUP DIR FILES POLICY =)**)
* Create file with name `delete_<filename>` and `<filename>` will be removed in `hot` and `backup` dirs
* Create file with name `delete_<RFC3339>_<filename>` and `<filename>` will be removed in `hot` and `backup` dirs at `<RFC3339>`

> example of `delete_<RFC3339>_<filename>`: `delete_2023-09-23T18:54:30+02:00_file_2`

## Basic usage

```bash
# for build
> make

# for run
> ./main

# for usage help
> ./main -h
```

## Backup mode

```bash
> ./main -backup-path ./backup -hot-path ./hot -log log.txt -log-level 0 -watcher syscall
```

* `-backup-path <path_to_backup_dir>`
* `-hot-path <path_to_origins_dir>`
* `-log <path_to_log_file>` - already existing log file will be appended
* `-log-level <number>` - `0` for Debug, `3` for Error
* `-watcher <type_of_watcher>`

> You could use default args if you want just to test it

## View logs mode

```bash
> ./main -view -view-date-filter "2023/09/25" -view-regex ".*CopyWorker.*"
```

* `-view` - switch to log viewing mode
* `-view-date-filter <date_to_search_for>` - format of the date should be same as in logs (`2009/01/23`)
* `--view-regex <some_regex>`

## Testing

```bash
# create example files to trigger them backup
> make create_files

# compare files in ./hot and ./backup dirs
> make compare_files

# create scheduled deletion of file_1
> make create_scheduled
```

Bash scripts used:
* `createFiles.sh` - creates given amount of files in given dir with names `file_<number>`
* `compareFiles.sh` - compares files in given dirs
* `scheduleDelete.sh` - creates file `delete_ISODATETIME_` for given file and sets date in filename for `now + 5 sec`

## Docker

If you don't have Linux, but you want to try this awesome app, try our simple Dockerfile!

```bash
# to build image
> make docker
# to run container and share current dir with container
> make docker-run
```

> Do not forget that Docker could create `./hot` and `./backup` dirs under `root`!

## Watchers

| Watcher | Description |
|-|-|
| syscal | Watcher that uses `syscall` with non-blocking IO. Fast, with small "loss window" (100ms), but could consume resources by triggering a lot of events |
| event | Watcher that uses `github.com/fsnotify/fsnotify`. Same `syscall` under the hood, produces more events than `syscall`, but more reliable |
| timer | Watcher that runs by ticker, uses much less resources, but has "loss window" (triggered once per sec) |

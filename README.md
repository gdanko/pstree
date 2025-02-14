# pstree
## Introduction
`pstree` is a small program that shows the process listing (`ps`) as a tree (as the name implies...). It has several options to make selection criteria and to change the output style. I ported the code from FredHucht's [repository](https://github.com/FredHucht/pstree) and added some more functionality.

It should compile under Linux, FreeBSD, OpenBSD, macOS, and Windows.

It uses [gopsutil](https://github.com/shirou/gopsutil) for gathering process information. I may use `ps` later so I can support more Un*x distirbutions, but for now, this is what you get.

## Compiling
* Clone this repository
* `cd` to the repository root
* Type `make build` and the binary will live under `bin` in the repository root
* You will need to manually copy `share/man/man1/pstree.1` to your MANPATH
* If you're using macOS, you can also use homebrew
    * `brew tap gdanko/homebrew`
    * `brew update`
    * `brew install gdanko/homebrew/pstree`

## Usage
```
$ pstree --help
Usage: pstree [-acUmntw] [--color] [-s, --contains <str>] [-l, --level <int>]
              [-g, --mode <int>] [-p, --pid <int>] [--rainbow] [-u, --user <str>]
   or: pstree -V

Display a tree of processes.

  -a, --arguments         show command line arguments
      --color             add some beautiful color to the pstree output
  -s, --contains string   show only branches containing process with <string> in commandline
  -c, --cpu               show CPU utilization percentage with each process, e.g., (c: 0.00%)
  -U, --exclude-root      don't show branches containing only root processes; cannot be used with --user
  -h, --help              help for pstree
  -l, --level int         print tree to <depth> level deep
  -m, --memory            show the memory usage with each process, e.g., (m: x.y MiB)
  -g, --mode int          use graphics chars for tree. n=1: IBM-850, n=2: VT100, n=3: UTF-8
  -n, --no-pids           do not show PIDs
  -p, --pid int32         show only branches containing process <pid>
      --rainbow           please don't
  -t, --threads           show the number of threads with each process, e.g., (t: xx)
  -u, --user string       show only branches containing processes of <user>; cannot be used with --exclude-root
  -V, --version           display version information
  -w, --wide              wide output, not truncated to window width

Process group leaders are marked with '='.
```

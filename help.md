| Short |                    Long                    |                            Purpose                             |
|-------|--------------------------------------------|----------------------------------------------------------------|
| -C,   | --alternate-rows                           | alternately change the line color                              |
|       | --caption string                           | custom caption                                                 |
| -i,   | --case-sensitive                           | case-sensitive in search                                       |
| -d,   | --column-delimiter character               | column delimiter character (default ",")                       |
| -c,   | --column-mode                              | column mode                                                    |
|       | --column-rainbow                           | column mode to rainbow                                         |
|       | --column-width                             | column mode for width                                          |
|       | --completion string                        | generate completion script [bash\|zsh\|fish\|powershell]       |
|       | --config file                              | config file (default is $XDG_CONFIG_HOME/ov/config.yaml)       |
|       | --debug                                    | debug mode                                                     |
|       | --disable-column-cycle                     | disable column cycling                                         |
|       | --disable-mouse                            | disable mouse support                                          |
| -e,   | --exec                                     | command execution result instead of file                       |
| -X,   | --exit-write                               | output the current screen when exiting                         |
| -a,   | --exit-write-after int                     | number after the current lines when exiting                    |
| -b,   | --exit-write-before int                    | number before the current lines when exiting                   |
|       | --filter string                            | filter search pattern                                          |
| -A,   | --follow-all                               | follow multiple files and show the most recently updated one   |
| -f,   | --follow-mode                              | monitor file and display new content as it is written          |
|       | --follow-name                              | follow mode to monitor by file name                            |
|       | --follow-section                           | section-by-section follow mode                                 |
| -H,   | --header int                               | number of header lines to be displayed constantly              |
| -h,   | --help                                     | help for ov                                                    |
|       | --help-key                                 | display key bind information                                   |
|       | --hide-other-section                       | hide other section                                             |
|       | --hscroll-width [int\|int%\|.int]          | width to scroll horizontally [int\|int%\|.int] (default "10%") |
|       | --incsearch[=true\|false]                  | incremental search (default true)                              |
| -j,   | --jump-target [int\|int%\|.int\|'section'] | jump target [int\|int%\|.int\|'section']                       |
| -n,   | --line-number                              | line number mode                                               |
|       | --memory-limit int                         | number of chunks to limit in memory (default -1)               |
|       | --memory-limit-file int                    | number of chunks to limit in memory for the file (default 100) |
| -M,   | --multi-color strings                      | comma separated words(regexp) to color .e.g. "ERROR,WARNING"   |
|       | --non-match-filter string                  | filter non match search pattern                                |
|       | --pattern string                           | search pattern                                                 |
| -p,   | --plain                                    | disable original decoration                                    |
| -F,   | --quit-if-one-screen                       | quit if the output fits on one screen                          |
|       | --regexp-search                            | regular expression search                                      |
|       | --section-delimiter regexp                 | regexp for section delimiter .e.g. "^#"                        |
|       | --section-header                           | enable section-delimiter line as Header                        |
|       | --section-header-num int                   | number of section header lines (default 1)                     |
|       | --section-start int                        | section start position                                         |
|       | --skip-extract                             | skip extracting compressed files                               |
|       | --skip-lines int                           | skip the number of lines                                       |
|       | --smart-case-sensitive                     | smart case-sensitive in search                                 |
| -x,   | --tab-width int                            | tab stop width (default 8)                                     |
| -v,   | --version                                  | display version information                                    |
|       | --view-mode string                         | apply predefined settings for a specific mode                  |
| -T,   | --watch seconds                            | watch mode interval(seconds)                                   |
| -w,   | --wrap[=true\|false]                       | wrap mode (default true)                                       |

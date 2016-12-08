csvbrowse
=========

On Windows, view CSV-File with the default web-browser.

```
[C:\] csvbrowse CSVPATH
```

- `%TEMP%\tmp.html` is created as a temporaly html file.
- And execute `cmd /C start %TEMP%\tmp.html`
- CSV-File is expected to be written as current code page(ANSI)
- When file's extension is tsv, change field seperater to TAB

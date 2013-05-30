ynote-import
============

Youdao note importer.

[Binary download](wiki/Binary)

Run with <code>-help</code> to see usage info:

```
Usage of ynote-import:
  yi [<flags>] [path] ...
Files are imported into the default folder. Files under a folder are imported to the corresponding folder, created if not exist. Subdirectories are not imported.
Options:
  -author="GO-IMPORTER": The author of imported notes.
  -enc="utf-8": The encoding of the input text.
  -reset=false: Reset to clean status. Forget saved access tokens.
  -sleep=0: Milliseconds to sleep after adding a note.
  -source="": The source of imported notes.
```

# Filed

Filed allows you to gather statistics about files on your system.
It stores the data in a local sqlite database for fast querying.

## Commands

### index -p "path to directory"

Indexes all files recursively rooted at "path to directory" 

### hash -d "path to sqlite database"

Hashes (sha512) all files in "path to sqlite database"

### stats -d "path to sqlite database"

Finds statistics of all hashed files in "path to sqlite database"
including 
* number of duplicates
* total file size spent

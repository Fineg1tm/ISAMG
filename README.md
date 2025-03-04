# ISAMG
Native Go ISAM file system (in initital development stage)

Master files definitions with one or more linked index files 
Fixed Length left or right justified data columns and specified padding, zeroes, spaces or none 
Fixed length data records terminated with CRLF or LF depending on the platform 
Marshal and Unmarshal functions using UTF8 Decode and Encode methods
File records are mapped in and out of master and index structs 
Struct datatypes currently supported are: string, int, float 
    Dates are stored as int in Julian date format CCYYDDD 
    Booleans can be used, but testing has been done with single character strings valued as 'Y' 'N'.

Create, Read, Update and Delete (CRUD) functionality has been tested
Transactional ROLLBACK, COMMIT is planned
Concurrency controls are implemented in the application level at present.

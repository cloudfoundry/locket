# How Locket is using database

Locket is using database to acquire component locks and update presences, which means that it will insert/update the locks table. 

The database table schema:

| table | column         | data type               | encrypted | description                                                                                                    |
|-------|----------------|-------------------------|-----------|----------------------------------------------------------------------------------------------------------------|
| locks | path           | character varying(255)  | NO        | Name of the lock                                                                                               |
|       | owner          | character varying(255)  | NO        | Bosh Job ID of the lock owner                                                                                  |
|       | value          | character varying(4096) | NO        | metadata set by the owner (only used by cells to store capacity information and available root-fs information) |
|       | type           | character varying(255)  | NO        | One of "lock" or "presence"                                                                                    |
|       | ttl            | bigint                  | NO        | Time to live (in seconds) of the lock                                                                          |
|       | modified_id    | character varying(255)  | NO        | GUID generated when the record is created                                                                      |
|       | modified_index | bigint                  | NO        | Integer incremented everytime there is an update to the record                                                 |

Locket client can define how frequently insert/update queries are performed. For both locks and presences client specifies retry interval and lock TTL. Locket client will try to acquire the lock or set the presence on specified interval. After the TTL is expired lock or presence will be removed from database.

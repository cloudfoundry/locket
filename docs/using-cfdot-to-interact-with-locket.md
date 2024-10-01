# Using cfdot to interact with Locket

cfdot command line tool provides the following commands to interact with Locket
server:

* `cfdot locks` - lists locks acquired in Locket. You can use this command to
  see the active BBS/Auctioneer when filtering logs on a given VM since the
  other instances are inactive.
* `cfdot claim-lock` - claims a Locket lock with the given key, owner, and
  optional value (run `cdfot claim-lock --help` for list of arguments).
* `cfdot presences` - lists presences registered in Locket
* `cfdot claim-presence` - claims a Locket presence with the given key, owner,
  and optional value (run `cdfot claim-presence --help` for list of arguments).

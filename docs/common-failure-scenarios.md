# Common failure scenarios

- `Context deadline exceeded` error in bbs logs. This means that the locket
  client (library in BBS) was not able to successfully claim a lock with the
  database. There could be networking, dns resolving, or database issues when
  such an event happens. Look for logs before `context deadline exceeded` for a
  RequestId to locket Server. Using the failed RequestId, we can then filter
  the locket server jobs to see more information. If the issue is still
  happening at the time of debugging, we can use the [`cfdot`
  cli](using-cfdot-to-interact-with-locket.md) to replicate the lock acquiring
  process.
- `Failed to acquire lock`. This message usually comes with error description
  in Locket client logs. 
   - `already acquired` error. That means that another component is holding the
     lock and current component will be in a passive mode until it acquires the
     lock. This error is considered an expected behavior.
   - `timeout` error. In that case the database load needs to be checked and
     communication between locket client, locket server and database needs to
     be verified. How long does it take to connect from locket server VM to
     database? What step is taking a long time: dial, dns resolution or server
     response? By default BBS locket client request timeout is 10s. It can be
     modified by changing value of `communication_timeout` in BBS.

## When BBS fails to acquire the lock

In the error scenario where BBS fails to acquire the lock the inactive BBS
should take over and acquire the lock. If the problem persists and all BBS
instances fail to acquire the lock there would be no active BBS instance in the
system which will result in that applications will be failing to stage and
start and new tasks failing to run. Existing applications and tasks should not
be affected unless there is evacuation happening at the same time.

## When Auctioneer fails to acquire the lock

When all instances of Auctioneer fail to acquire the lock applications staging,
staring and new tasks running will be failing. Existing applications and tasks
should not be affected unless there is evacuation happening at the same time.
Auctioneer will always have a single active instance (like BBS), so if one
fails to acquire lock, another one will try to acquire instead and become
active.

## When Rep fails to update its presence

If Rep is failing to update its presence, BBS will mark all LRPs on that Rep as
Suspect and it will start rescheduling them on active Reps. This might result
in resource contention (insufficient resources error when pushing an
application) and re-shuffling the LRPs.



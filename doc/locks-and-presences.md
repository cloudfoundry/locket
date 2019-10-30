# Locks And Presences

There are to types of entities that are managed by Locket:

* Locks - only one lock can be acquired by the lock name at a time. The
  component that is holding the lock is considered an active component. Then it
  starts it tries to acquire the lock before serving any request or performing
  periodic operations. The following components are using Locket to acquire the
  lock in order to pick an active component:
  * BBS (retry interval: 5s, ttl: 15s)
  * Auctioneer (retry interval: 5s, ttl: 15s)
 
* Presences - updated periodically to register component as active and being
  able to serve requests. Each Diego rep component is updating its presence
  with Locket. Retry interval: 5s TTL: 15s. Auctioneer will use the information
  provided by presences (by querying the database) to determine which Rep
  components are available to run the LRP/Task. When Rep does not update, it
  looses its presence and its LRPs are marked as suspect by BBS and they will
  start to be rescheduled to run on active Reps. If the Rep becomes active
  again all suspect LRPs will become active again and rescheduled LRPs will be
  deleted.

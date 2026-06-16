---
title: Understanding Locket Logs
expires_at : never
tags: [diego-release, locket]
---

# Understanding Locket Logs

This doc will go through a few typical examples using locket and show you the log output generated from these scenerios.

## "Lock" locks
* I have two VMs where BBS is deployed
  * `control/9e859b27-ac41-4e39-ba59-bb87d520d987` 👈 This VM started first. So the bbs on it has the lock.
  * `control/fe89652e-4e59-4efc-b75c-88cc029ed816` 👈 This VM started second. So the bbs on it does not have the lock.
 
 ```
# control/9e859b27-ac41-4e39-ba59-bb87d520d987
# bbs.stdout.log

# This BBS has the lock, so it is active.
# It is constantly logging about the work it is doing.
...
{"timestamp":"2023-06-22T19:13:05.362719089Z","level":"info","source":"bbs","message":"bbs.converge-lrps.db-converge-lrps.complete","data":{"session":"1875.2"}}
{"timestamp":"2023-06-22T19:13:05.366102972Z","level":"info","source":"bbs","message":"bbs.executing-convergence.converge-lrps-done","data":{"session":"1874"}}
...
 ```

 ```
# control/fe89652e-4e59-4efc-b75c-88cc029ed816
# bbs.stdout.log

# This BBS does not have the lock, so it is not active.
# This last log line is from when it attempted to grab the lock, but failed because a different bbs held the lock.
# At this point it is not actively logging because it is not doing any work.
# Note that it tries to make a lock with owner 'fe89652e-4e59-4efc-b75c-88cc029ed816' (this VM), 
# but the lock already exists with owner '9e859b27-ac41-4e39-ba59-bb87d520d987' (the other VM with the bbs that has the lock).

{"timestamp":"2023-06-22T18:01:32.238648111Z","level":"error","source":"bbs","message":"bbs.locket-lock.failed-to-acquire-lock",
"data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","lock":{"key":"bbs","owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","type":"lock","type_code":1},
"lock-owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","request-uuid":"7ca981d9-9375-45a8-7ddc-8940dc662643","session":"4","ttl_in_seconds":15}}
```

```
# locket.stdout.log

# There is a matching log in locket from where the bbs on control/fe89652e-4e59-4efc-b75c-88cc029ed81
attempted to grab the lock and failed.

{"timestamp":"2023-06-22T18:01:32.238648111Z","level":"error","source":"bbs","message":"bbs.locket-lock.failed-to-acquire-lock",
"data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","lock":{"key":"bbs","owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","type":"lock","type_code":1},
"lock-owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","request-uuid":"7ca981d9-9375-45a8-7ddc-8940dc662643","session":"4","ttl_in_seconds":15}}
 ```

```
# In the locket.lock db table there is one lock for bbs.
# And the owner is the VM guid of the VM where the bbs that has the lock is running.
# The 'modified_index' is constantly being sequentially updated.

mysql> select * from locket.locks where path = "bbs";
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
| path | owner                                | value | type | modified_index | modified_id                          | ttl  |
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
| bbs  | 9e859b27-ac41-4e39-ba59-bb87d520d987 |       | lock |           4352 | d6d8b4f2-8d1e-4ba0-71ca-fdbf6b0bad58 |   15 |
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
```

### Changing who has locks
If the bbs that currently has the lock restarts, then it will give up the lock and the other bbs will claim the lock.
I did this by running `monit restart bbs` on `control/9e859b27-ac41-4e39-ba59-bb87d520d987`.

```
# locket.stdout.log
# The BBS on VM '9e859b27-ac41-4e39-ba59-bb87d520d987' gives up the lock.
{"timestamp":"2023-06-22T19:16:05.620930708Z","level":"info","source":"locket","message":"locket.release.release-lock.released-lock","data":{"key":"bbs","owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","session":"1774728.1","type":"lock","type-code":1}}

# The BBS on VM 'fe89652e-4e59-4efc-b75c-88cc029ed816' successfully claims the lock.
{"timestamp":"2023-06-22T19:16:06.089322785Z","level":"info","source":"locket","message":"locket.lock.lock.acquired-lock","data":{"key":"bbs","owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","request-uuid":"7e2a16d0-56a7-40dc-5d04-ca7275364cc2","session":"1774735.1","type":"lock","type-code":1}}

# The BBS on VM '9e859b27-ac41-4e39-ba59-bb87d520d987' is restarted and attempts to claim the lock.
{"timestamp":"2023-06-22T19:16:20.492246371Z","level":"info","source":"locket","message":"locket.lock.register-ttl.fetch-and-release-lock.fetched-lock","data":{"key":"bbs","modified-index":4431,"owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","request-uuid":"05c37d7c-4163-4b64-4d9f-30af73a0860d","session":"1774724.2.1","type":"lock","type-code":1}}

# The BBS on VM '9e859b27-ac41-4e39-ba59-bb87d520d987' fails
# to claim the lock, because the bbs on VM 'fe89652e-4e59-4efc-b75c-88cc029ed816' has the lock.
{"timestamp":"2023-06-22T19:16:20.493343559Z","level":"error","source":"locket","message":"locket.lock.register-ttl.fetch-and-release-lock.fetch-failed-owner-mismatch","data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","fetched-owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","key":"bbs","modified-index":4431,"owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","request-uuid":"05c37d7c-4163-4b64-4d9f-30af73a0860d","session":"1774724.2.1","type":"lock","type-code":1}}
{"timestamp":"2023-06-22T19:16:20.495265778Z","level":"error","source":"locket","message":"locket.lock.register-ttl.failed-compare-and-release","data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","key":"bbs","modified-index":4431,"request-uuid":"05c37d7c-4163-4b64-4d9f-30af73a0860d","session":"1774724.2","type":"lock"}}
 ```

```
# The lock in the DB is now updated with the new owner.

mysql> select * from locket.locks where path = "bbs";
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
| path | owner                                | value | type | modified_index | modified_id                          | ttl  |
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
| bbs  | fe89652e-4e59-4efc-b75c-88cc029ed816 |       | lock |             98 | cba003ea-608f-4cdf-539e-0304896dc313 |   15 |
+------+--------------------------------------+-------+------+----------------+--------------------------------------+------+
1 row in set (0.00 sec)
```

```
# control/9e859b27-ac41-4e39-ba59-bb87d520d987
# bbs.stdout.log

# This BBS has successfully restarted.
{"timestamp":"2023-06-22T19:16:07.128143147Z","level":"info","source":"bbs","message":"bbs.locket-lock.started","data":{"lock":{"key":"bbs","owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","type":"lock","type_code":1},"session":"4","ttl_in_seconds":15}}

# This BBS fails to claim the lock because the other bbs has the lock
{"timestamp":"2023-06-22T19:16:07.136452192Z","level":"error","source":"bbs","message":"bbs.locket-lock.failed-to-acquire-lock","data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","lock":{"key":"bbs","owner":"9e859b27-ac41-4e39-ba59-bb87d520d987","type":"lock","type_code":1},"lock-owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","request-uuid":"efb3d6c6-9585-4efb-6f74-126dcfccb99c","session":"4","ttl_in_seconds":15}}
```

```
# control/fe89652e-4e59-4efc-b75c-88cc029ed816
# bbs.stdout.log

# This BBS grabs the lock as soon as it is up for grabs.
{"timestamp":"2023-06-22T19:16:06.091053490Z","level":"info","source":"bbs","message":"bbs.locket-lock.acquired-lock","data":{"lock":{"key":"bbs","owner":"fe89652e-4e59-4efc-b75c-88cc029ed816","type":"lock","type_code":1},"session":"4",
"ttl_in_seconds":15}}

# Once it grabs the lock it gets to work!
{"timestamp":"2023-06-22T19:16:06.092710053Z","level":"info","source":"bbs","message":"bbs.set-lock-held-metron-notifier.started","data":{"session":"5"}}
{"timestamp":"2023-06-22T19:16:06.093315811Z","level":"info","source":"bbs","message":"bbs.task-completion-workpool.starting","data":{"session":"1"}}
...
```
## "Presence" locks

Presence locks are registered by the rep on each diego cell.
It puts information about the diego cell into the lock db so BBS can know which diego cells are available to run workloads.

I have two diego cells with GUIDs: `0e776c46-fd44-48f4-83ea-06b1dd13ca4b` and `8ba76ac6-4809-4255-b59b-2e3878bf23d1`.

```
# locket.lock db table

# There is one presence lock per diego cell, where the path is the GUID of the diego cell VM.

mysql> select path, type from locks;
+--------------------------------------+----------+
| path                                 | type     |
+--------------------------------------+----------+
| 0e776c46-fd44-48f4-83ea-06b1dd13ca4b | presence |
| 8ba76ac6-4809-4255-b59b-2e3878bf23d1 | presence |
| auctioneer                           | lock     |
| bbs                                  | lock     |
| cc-deployment-updater                | lock     |
| policy-server-asg-syncer             | lock     |
| routing_api_lock                     | lock     |
| tps_watcher                          | lock     |
+--------------------------------------+----------+
8 rows in set (0.00 sec)
```

## Changing the presence locks
1. run `monit retstart rep` on diego cell `0e776c46-fd44-48f4-83ea-06b1dd13ca4b`. This simulates an upgrade situation where the rep is restarted.

```
# rep.stdout.log

# Rep exits
{"timestamp":"2023-06-22T20:19:38.920038544Z","level":"info","source":"rep","message":"rep.exited","data":{}}

# Rep restarts 
{"timestamp":"2023-06-22T20:19:40.263986195Z","level":"info","source":"rep","message":"rep.wait-for-garden.ping-garden","data":{"initialTime:":"2023-06-22T20:19:40.263953792Z","session":"1","wait-time-ns:":30876}}

...
# Rep gets the lock again
{"timestamp":"2023-06-22T20:19:42.926917888Z","level":"info","source":"rep","message":"rep.locket-lock.acquired-lock","data":{"lock":{"key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","owner":"c7ed4409-d33a-4b9b-644c-e5cdd873568f","value":"{
\"cell_id\":\"0e776c46-fd44-48f4-83ea-06b1dd13ca4b\",\"rep_address\":\"http://10.0.4.8:1800\",\"zone\":\"us-central1-f\",\"capacity\":{\"memory_mb\":12977,\"disk_mb\":104349,\"containers\":249},\"rootfs_provider_list\":[{\"name\":\"prel
oaded\",\"properties\":[\"cflinuxfs3\",\"cflinuxfs4\"]},{\"name\":\"preloaded+layer\",\"properties\":[\"cflinuxfs3\",\"cflinuxfs4\"]},{\"name\":\"docker\"}],\"rep_url\":\"https://0e776c46-fd44-48f4-83ea-06b1dd13ca4b.cell.service.cf.inte
rnal:1801\"}","type":"presence","type_code":2},"session":"12","ttl_in_seconds":15}}
```

```
# locket.stdout.log

# The lock is released when Rep is restarted.
{"timestamp":"2023-06-22T20:19:38.914625536Z","level":"info","source":"locket","message":"locket.release.release-lock.released-lock","data":{"key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","owner":"78e5a54a-cc4f-4f67-5efc-6c2309bd5d28","session":"1815120.1","type":"presence","type-code":2}}

# The lock is acquired again.
{"timestamp":"2023-06-22T20:19:42.926185907Z","level":"info","source":"locket","message":"locket.lock.lock.acquired-lock","data":{"key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","owner":"c7ed4409-d33a-4b9b-644c-e5cdd873568f","request-uuid":"3293a3eb-3c86-4965-457b-3c54cd3c4199","session":"1815164.1","type":"presence","type-code":2}}

# For some reason it tries to grab the lock again (?) but can't (?).
# This confusing error seems to be a part of the normal process.
{"timestamp":"2023-06-22T20:19:53.862699986Z","level":"info","source":"locket","message":"locket.lock.register-ttl.fetch-and-release-lock.fetched-lock","data":{"key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","modified-index":91,"owner":"78e5a54a-cc4f-4f67-5efc-6c2309bd5d28","request-uuid":"733a2b67-f984-46ff-4f04-59348b6e47d7","session":"1815119.2.1","type":"presence","type-code":2}}
{"timestamp":"2023-06-22T20:19:53.862867442Z","level":"error","source":"locket","message":"locket.lock.register-ttl.fetch-and-release-lock.fetch-failed-owner-mismatch","data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","fetched-owner":"c7ed4409-d33a-4b9b-644c-e5cdd873568f","key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","modified-index":91,"owner":"78e5a54a-cc4f-4f67-5efc-6c2309bd5d28","request-uuid":"733a2b67-f984-46ff-4f04-59348b6e47d7","session":"1815119.2.1","type":"presence","type-code":2}}
{"timestamp":"2023-06-22T20:19:53.863737078Z","level":"error","source":"locket","message":"locket.lock.register-ttl.failed-compare-and-release","data":{"error":"rpc error: code = AlreadyExists desc = lock-collision","key":"0e776c46-fd44-48f4-83ea-06b1dd13ca4b","modified-index":91,"request-uuid":"733a2b67-f984-46ff-4f04-59348b6e47d7","session":"1815119.2","type":"presence"}}

```

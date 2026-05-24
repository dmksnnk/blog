---
date: '2026-05-24T14:00:00+02:00'
draft: false
title: 'CRDT and Event-Driven Systems'
slug: 'crdt-and-event-driven-systems'
showToc: true
cover:
    image: 'images/crdt.svg'
summary: |
  Conflict-free Replicated Data Types and event-driven systems are
  closely related, let's explore it.
tags:
    - crdt
    - event-driven
    - distributed systems
---

# CRDT and Event-Driven Systems

Conflict-free Replicated Data Type. Mostly known for its use in collaborative
software, like collaborative editing tools. The name CRDT is a little bit
misleading, not that there are no conflicts, just that conflict resolution is
built into the data type itself.

We will look at them from the side of event-driven distributed systems.
You will see that those are closely related and lessons from CRDTs directly
applicable to event-driven systems.

## Total order

As the name implies, this is when we know the total order of events.
For example, in the bank account, we have a log of all deposits and withdrawals:

```
+$100 -$80 +$20 -$40
```
If we change the order of events, we'll get overdraft:

```
+$100 -$80 -$40 = -$20, oops, overdraft!
```

Total order is a time-ordered log of all changes, from creation of the
object to its death.

In databases, each transaction is an event. In serializable isolation
level (the highest one), database does different kind of tricks to
guarantee that each transaction is executed in such a way,
that appear to be executed in serial order, one after another.
This means transactions are in total order.

## Partial order

Causal order - _happen before_ - one event happened before the other,
we do not have or do not need the _total order_ of all events.
Causality just means that you know what caused what.

For example, we have user login-logout events in Kafka topic.
`user1` events are always stored in partition 1, `user2` events are stored
in partition 2. Events for `user1` are _causally ordered_, e.g. logout cannot
happen before login:

```
partition 1: (user1 login)                               (user1 logout)
partition 2:                (user2 login) (user2 logout)

```

Events for `user1` and `user2` are _concurrent_. Events for `user2` may or
may not happen before events for the `user1` _in total order_, and, frankly,
we don’t care. We also may process events not in the total order,
e.g. process `user1` login after `user2` login.

When we do not care about total order of events,
this is known as _optimistic replication_.

## Commutative operations

Sometimes, you don’t even need a partial order. For commutative operations,
like adding numbers, the order is not important.

### Counter

A good example is a distributed counter. Imagine a like button on Youtube video.
Every time a user clicks +1, an event is sent. To calculate total number of likes,
you just count all the click events:

```
+1 +1 +1 ... +1 => aggregate
```

There can be multiple levels of aggregation:

```
+1 +1 +1 ... +1 => aggregate => +5 +7 ...\
                                           => aggregate => +20 +37 ...
+1 +1 +1 ... +1 => aggregate => +2 +1 .../
```

In whatever order we process the events, result will be the same.

The same works for dislikes: you have one positive counter and one negative
counter, the total count will be positive count minus negative count.

**Grow-Only Counter** is the first CRDT we have encountered.

### Max-Register

Imagine an auction system, where bids arrive from many clients across distributed
replicas. Each replica tracks the highest bid for the item.
Whenever a new bid arrives, we compare it with the current highest bid,
if it is higher - we update the highest bid, if not - ignore it.

```
ReplicaA -> $100, $50, $120 => max => $120  \
                                             => max => $160
ReplicaB -> $110, $130, $160 => max => $160 /
```

This is known as **Max-Register**. The `max` function is commutative - the order
of operations does not affect the result.

### Grow-only Set

Back to our Youtube video, we want to track all unique users who watched the
video. We can have a set of user IDs and whenever we receive an event,
we check if user ID is already in the set, if not - we add it, if yes -
ignore it.

```
ReplicaA -> 1, 2, 1 => set => {1, 2}   \
                                         => set => {1, 2, 3, 4}
ReplicaB -> 2, 3, 4 => set => {2, 3, 4} /
```

This is known as **Grow-only Set**. It only allows adds, once added, element
cannot be removed.

## Restoring order of events

In partially ordered systems, it is still possible to get the order of
events after the fact. It is done by using _logical clocks_. The common choice is
[Lamport timestamp](https://en.wikipedia.org/wiki/Lamport_timestamp)
for global ordering or [Vector clock](https://en.wikipedia.org/wiki/Vector_clock)
for causality between events.

When a system produces an event, it adds a logical timestamp to all events -
always increasing value. This can be a simple counter for all produced events,
so each new produced event has `timestamp = timestamp + 1`.

> [!NOTE]
> Using actual clock for timestamps is a bad idea, clocks are often out-of-sync,
> two computers in the same room can have difference in tens of minutes.
> Imagine processing tens of thousands of events per second,
> microsecond drift of the clock may cause big problems.


Let’s say we have processed events in the order:

```
(1),(3),(2),(4)
```

Restoring the total order is fairly simple: order events by timestamp.

```
(1),(2),(3),(4)
```


### Conflict resolution

When events are produced from multiple replicas, timestamp often includes
replica ID for conflict resolution. Imagine two replicas updating users
`user1` and `user2` simultaneously and producing events on each update.
Each event has a logical timestamp `(counter, replicaID)`:

```
ReplicaA -> user1(1,A),user2(2,A),user1(3,A)
ReplicaB -> user1(2,B),user2(3,B),user1(4,B)
```

Events are ordered by the counter part of the timestamp, replicaID is used to
resolve the conflict on ties. Total order will be:

```
user1(1,A),user2(2,A),user1(2,B),user1(3,A),user2(3,B),user1(4,B)
```

This conflict resolution is known as **Last-Write-Wins (LWW)**.

### Types of CRDTs

The problem can happen when events are not just a log, which can be reordered,
but changes applied to the object.

Looking back at our user changes events example. If user changes
its username, system can send only change, for example
`{userId: 123, username: biggus-dickus-69}`
or you can send the whole state of the user object
`{userId: 123, username: biggus-dickus-69, email: ..., phone: ...}`.

When you are providing only changes, it is called **Operation-Based CRDT**.
When you are distributing whole state, it is called **State-Based CRDT**.
Both have their pros and cons. Operation-Based CRDTs are smaller,
as you propagate only the change, but have more complex conflict resolution.
State-Based CRDTs are bigger, but conflict resolution is simpler.

Imagine, you have a system which reads user changes from ReplicaA and ReplicaB,
and applies these changes to an object. For a short amount of time,
you have lost a connection for ReplicaB and was processing events only
from ReplicaA. You have processed `e1(1,A),e2(2,A),e1(3,A)`, then connection
got fixed and you are processing `e1(2,B),e2(3,B),e1(4,B)`.

```
e1(1,A),e2(2,A),e1(3,A)

---- broken connection ----  e1(2,B),e2(3,B),e1(4,B)
```

You need to restore the order of events to get to the correct state of the user.

In the case of state-based CRDTs (whole state propagated),
you just need to process events that have a bigger timestamp than you already have.
For our example, we already have entities in the version `e1(3,A),e2(2,A)`,
so we can skip events with lower timestamps and process only `e2(3,B),e1(4,B)`.

```
e1 has a version (3,A)
e1(3,A) > e1(2,B) => e1(3,A) // skip processing
e1(3,A) < e1(4,B) => e1(4,B) // apply

e2 has a version (2,A)
e2(2,A) < e2(3,B) => e2(3,B) // apply
```

In the case of operation-based CRDTs (only changes propagated), things get more
complex. The solution is to undo all events up to `e1(1,A)` and re-apply events
in correct order. This approach is known as a _time-warp_.

```
e1(1,A),e2(2,A),e1(3,A)──┐
e1(1,A),e2(2,A) ◀──undo──┘
e1(1,A)
  ┌─redo──┐
  │       ▼
e1(1,A),e1(2,B)
e1(1,A),e1(2,B),e2(2,A)
e1(1,A),e1(2,B),e2(2,A),e1(3,A)
e1(1,A),e1(2,B),e2(2,A),e1(3,A),e2(3,B)
e1(1,A),e1(2,B),e2(2,A),e1(3,A),e2(3,B),e1(4,B)
```

This process can be expensive, as we need to store the different states of the
object to be able to undo. Also, this can have unexpected _external_ side
effects. For example, we had a state that tells the system to send an email
about successful booking of a flight, but later, we receive an event,
that happened earlier, telling that the flight was actually cancelled before the
booking happened. Now we need a "compensation" action, e.g. send an apology
email explaining a situation, which may or may not be acceptable depending on
the use-case.

## Sequence CRDTs

All of the above also applies to collaborative editors.
Each change (character added, item in the list moved, etc.) is an event.
All events are stored in a sequence, which is the same as an event log.
When several users edit the document offline and then the document is synced,
editor receives sequences (log of events) from the users and merge them together
to receive **total order of events**.

For example, we have a document with a “Hello World” text in it. Each symbol in
a text got assigned an index. These indexes are constant and do not change:

```
 H   e   l   l   o   _   W   o   r   l    d
1.0 2.0 3.0 4.0 5.0 6.0 7.0 8.0 9.0 10.0 11.0
```

Let’s say, there are two users, A and B, both editing document offline.
User A adds an exclamation mark `!` at the end. It got assigned a new index:

```
 H   e   l   l   o   _   W   o   r   l    d    !
1.0 2.0 3.0 4.0 5.0 6.0 7.0 8.0 9.0 10.0 11.0 12.0
```

At the same time, user B adds a comma `,` after `o`, it gets assigned a new
index number:

```
 H   e   l   l   o   ,   _   W   o   r   l    d
1.0 2.0 3.0 4.0 5.0 5.5 6.0 7.0 8.0 9.0 10.0 11.0
```

Synchronization happens, both editors receive each other's sequences and
merge the changes:

```
 H   e   l   l   o   ,   _   W   o   r   l    d    !
1.0 2.0 3.0 4.0 5.0 5.5 6.0 7.0 8.0 9.0 10.0 11.0 12.0
```

In this example, we are using float numbers, so we can get very large
number of in-between indexes. In practice, variable-depth integers are used,
like [LOGOOT](https://inria.hal.science/inria-00432368/document/) and
[LSEQ](https://hal.science/hal-00921633/document).

There are different collaborative editors, under the hood they differ
in how they resolve conflicts and store sequences.

## Conclusion

The world of CRDTs and collaborative editors is very close to event-based
distributed systems.

- If you look hard enough, event-driven services are kind-of eventual
  consistent replicas of each other, with different views on the same data.
- Try to avoid conflicts, use _commutative operations_ where possible.
- Use _version clocks_ to get a causal order of events per entity if
  order is important.

## Sources

- [Thinking in events](https://dl.acm.org/doi/10.1145/3465480.3467835)
- [Lamport timestamp](https://en.wikipedia.org/wiki/Lamport_timestamp)
- [Vector clock](https://en.wikipedia.org/wiki/Vector_clock)
- [Version Vectors are not Vector Clocks](https://haslab.wordpress.com/2011/07/08/version-vectors-are-not-vector-clocks/)
- [CRDTs: The Hard Parts](https://www.youtube.com/watch?v=x7drE24geUw)

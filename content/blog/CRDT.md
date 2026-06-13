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

Conflict-free Replicated Data Types (CRDTs) are mostly known for their use in
collaborative software, like collaborative editing tools. The name CRDT
is a little bit misleading: not that there are no conflicts, just that
conflict resolution is built into the data type itself.

We will look at them from the perspective of event-driven distributed systems.
You will see that they are closely related and that lessons from CRDTs are
directly applicable to event-driven systems. First, some definitions.

## Total order

As the name implies, this is when we know the total order of events.
For example, in a bank account, we have an *ordered* log of
all deposits and withdrawals:

```
+$100 -$80 +$20 -$40
```
The order is important here: if we change the order of events,
we'll get an overdraft:

```
+$100 -$80 -$40 = -$20, oops, overdraft!
```

A total order is a time-ordered log of all changes, from the creation of the
object to its death. In this case, it can be a log of all operations
from the opening of the bank account to its closing.

In databases, each transaction is an event. At the serializable isolation
level (the highest one), a database uses different kinds of tricks to
guarantee that transactions appear to be executed in serial order,
one after another.
This means transactions are in total order.

## Partial order

Causal order - _happened before_ - means one event happened before another,
we do not have or do not need the _total order_ of all events.
Causality just means that you know what caused what.

For example, we have user login-logout events in a Kafka topic.
`user1` events are always stored in partition 1, `user2` events are stored
in partition 2. Events for `user1` are _causally ordered_, e.g. logout cannot
happen before login:

```
partition 1: (user1 login)                               (user1 logout)
partition 2:                (user2 login) (user2 logout)

```

Events for `user1` and `user2` are _concurrent_. Events for `user2` may or
may not happen before events for `user1` _in total order_, and, frankly,
we don’t care. We also may process events not in the total order,
e.g. process `user1` login after `user2` login.

When we do not care about the total order of events,
it is known as _optimistic replication_.

## Commutative operations

Sometimes, you don’t even need a partial order. For commutative operations,
like adding numbers, the order is not important.

### Counter

A good example is a distributed counter. Imagine a like button on a YouTube video.
Every time a user clicks +1, an event is sent. To calculate the total number of likes,
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

In whatever order we process the events, the result will be the same.

The same works for dislikes: you have one positive counter and one negative
counter, and the total count will be the positive count minus the negative count.

**Grow-Only Counter** is the first CRDT we have encountered.

### Max-Register

Imagine an auction system, where bids arrive from many clients across distributed
replicas. Each replica tracks the highest bid for the item.
Whenever a new bid arrives, we compare it with the current highest bid.
If it is higher, we update the highest bid, if not, we ignore it.

```
ReplicaA -> $100, $50, $120 => max => $120  \
                                             => max => $160
ReplicaB -> $110, $130, $160 => max => $160 /
```

This is known as **Max-Register**. The `max` function is commutative - the order
of operations does not affect the result.

### Grow-only Set

Back to our YouTube video, we want to track all unique users who watched the
video. We can have a set of user IDs and whenever we receive an event,
we check if the user ID is already in the set. If not, we add it, if yes -
ignore it.

```
ReplicaA -> 1, 2, 1 => set => {1, 2}   \
                                         => set => {1, 2, 3, 4}
ReplicaB -> 2, 3, 4 => set => {2, 3, 4} /
```

This is known as **Grow-only Set**. It only allows adds, once added, an element
cannot be removed.

## Restoring order of events

In partially ordered systems, it is still possible to get the order of
events after the fact. This is done by using _logical clocks_. The common choices are
[Lamport timestamp](https://en.wikipedia.org/wiki/Lamport_timestamp)
for global ordering or [Vector clock](https://en.wikipedia.org/wiki/Vector_clock)
for causality between events.

When a system produces an event, it adds a logical timestamp to it -
an always-increasing value. This can be a simple counter for all produced events,
so each new produced event has `timestamp = timestamp + 1`.

{{< alert note >}}

Using an actual clock for timestamps is a bad idea. Clocks are often out of sync,
two computers in the same room can have a difference of tens of minutes.
Imagine processing tens of thousands of events per second:
a microsecond drift of the clock may cause big problems.

{{< /alert >}}

Let’s say we have processed events in the order:

```
(1),(3),(2),(4)
```

Restoring the total order is fairly simple: order events by timestamp.

```
(1),(2),(3),(4)
```


### Conflict resolution

When events are produced from multiple replicas, the timestamp often includes
replica ID for conflict resolution. Imagine two replicas updating users
`user1` and `user2` simultaneously and producing events on each update.
Each event has a logical timestamp `(counter, replicaID)`:

```
ReplicaA -> user1(1,A),user2(2,A),user1(3,A)
ReplicaB -> user1(2,B),user2(3,B),user1(4,B)
```

Events are ordered by the counter component of the timestamp, and replicaID is used to
resolve the conflict on ties.

- At timestamp `1` we have only one event `user1(1,A)`, so it is the first event.
- At timestamp `2` we have two events `user2(2,A)` and `user1(2,B)`, so we need to
  decide which one goes first. We can use replica ID to resolve the conflict:
  `A` goes before `B`, so `user2(2,A)` goes before `user1(2,B)`.
- At timestamp `3` there is a similar situation, we resolve the conflict by replica ID,
  so `user1(3,A)` goes before `user2(3,B)`.

Total order will be:

```
user1(1,A),user2(2,A),user1(2,B),user1(3,A),user2(3,B),user1(4,B)
```

This conflict resolution is known as **Last-Write-Wins (LWW)**.

### Types of CRDTs

The problem can happen when events are not just a log, which can be reordered,
but changes that are applied to the object.

Looking back at our user change events example. If a user changes
their username, the system can send only the change, for example
`{userId: 123, username: biggus-dickus-69}`
or you can send the whole state of the user object
`{userId: 123, username: biggus-dickus-69, email: ..., phone: ...}`.

When you are providing only changes, it is called an **Operation-Based CRDT**.
When you are distributing the whole state, it is called a **State-Based CRDT**.
Both have their pros and cons. Operation-Based CRDTs are smaller,
as you propagate only the change, but have more complex conflict resolution.
State-Based CRDTs are bigger, but conflict resolution is simpler.

Imagine you have a system that reads user changes from ReplicaA and ReplicaB
and applies these changes to an object. For a short amount of time,
you lost a connection to ReplicaB and were processing events only
from ReplicaA. You processed `e1(1,A),e2(2,A),e1(3,A)`, then the connection
got fixed and you are processing `e1(2,B),e2(3,B),e1(4,B)`.

```
e1(1,A),e2(2,A),e1(3,A)

---- broken connection ----  e1(2,B),e2(3,B),e1(4,B)
```

You need to restore the order of events to get to the correct state of the user.

In the case of State-Based CRDTs (whole state propagated), you just need to
process events that have a bigger timestamp than you already have.
For our example, before the connection is restored, we have entities at
version `e1(3,A),e2(2,A)`, so we can skip events with lower timestamps and
process only `e2(3,B),e1(4,B)`.

```
e1 has a version (3,A)
e1(3,A) > e1(2,B) => e1(3,A) // skip processing
e1(3,A) < e1(4,B) => e1(4,B) // apply

e2 has a version (2,A)
e2(2,A) < e2(3,B) => e2(3,B) // apply
```

In the case of Operation-Based CRDTs (only changes propagated), things get more
complex. The solution is to undo all events up to `e1(1,A)` and re-apply events
in the correct order. This approach is known as a _time-warp_.

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
about the successful booking of a flight, but later, we receive an event
that happened earlier, telling us that the flight was actually cancelled before the
booking happened. Now we need a "compensation" action, e.g. send an apology
email explaining the situation, which may or may not be acceptable depending on
the use-case.

## Sequence CRDTs

All of the above also applies to collaborative editors.
Each change (a character added, an item in the list moved, etc.) is an event.
All events are stored in a sequence, which is the same as an event log.
When several users edit the document offline and then the document is synced,
the editor receives sequences (logs of events) from the users and merges them together
to receive a **total order of events**.

For example, we have a document with “Hello World” text in it. Each symbol in
the text gets assigned an index. These indexes are constant and do not change:

```
 H   e   l   l   o   _   W   o   r   l    d
1.0 2.0 3.0 4.0 5.0 6.0 7.0 8.0 9.0 10.0 11.0
```

Let’s say, there are two users, A and B, both editing document offline.
User A adds an exclamation mark `!` at the end. It gets assigned a new index:

```
 H   e   l   l   o   _   W   o   r   l    d    !
1.0 2.0 3.0 4.0 5.0 6.0 7.0 8.0 9.0 10.0 11.0 12.0
```

At the same time, user B adds a comma `,` after `o`, it gets assigned a new
index between `5.0` and `6.0`:

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

In this example, we are using floating-point numbers, so we can get a very large
number of in-between indexes. In practice, variable-depth integers are used,
like [LOGOOT](https://inria.hal.science/inria-00432368/document/) and
[LSEQ](https://hal.science/hal-00921633/document).

There are different collaborative editors, under the hood they differ
in how they resolve conflicts and store sequences.

## Conclusion

The world of CRDTs and collaborative editors is very close to event-based
distributed systems.

- If you look hard enough, event-driven services are kind of eventually
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

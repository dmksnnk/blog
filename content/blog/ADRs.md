---
date: '2025-08-14T10:10:00+02:00'
title: 'Making technical decisions'
slug: 'decision-records'
showToc: true
draft: false
cover:
    image: images/decision-records.svg
summary: "Making better technical decisions with decision records."
tags:
    - ADR
    - architecture
---

## Making Technical Decisions

TL;DR: Write down the options for solving the problem, with the pros and cons of each.
Discuss. Write down how the actual solution will be implemented.

After reading about the writing culture and the use of ADRs at other companies, we decided to try it out.
So far, we have been able to confirm:
- **It spreads knowledge**: Instead of answering the same questions again and again,
    you can write down the answers once and redirect people to the document with all the relevant information.
- **Consensus**: This should be a place where all team members are free to express their perspectives and concerns
    and to find the best solution given the requirements and limitations.
- **It preserves knowledge**: I often find myself going back to it to remember what we chose to implement and why.
- **Paper trail**: It can be evidence of what was agreed upon, to cover you if something goes south.
- **Time well invested**: Many corner cases and potential problems are discovered long before the actual implementation.
    Often, implementing and writing the code is the easiest part.

Most importantly, I was able to confirm a well-known fact:
writing down ideas helps to find the holes in them, even when the idea seems perfect in your head.

If a decision is important and could have significant consequences, every involved party should
read the document and give their yea or nay. Sometimes it is hard to make them read it because everyone has their own
stuff to care about. For example, at AWS, for the first 10-15 minutes of a meeting, people sit in silence
and read the whole document. After that, you can have a productive discussion, as everybody
is on the same page. This might be a bit awkward. Instead, we just jump on a call/meeting with
all involved parties to quickly go through and explain the document.

Below is an overview of how we implement the technical decision process.
Ideally, it should be simplified even more, but simple is hard. We are trying to remove as much friction as possible.
Your process may differ, but this is what has worked for us.

## High-level process

Let's take a look at the process as a whole:

1. First, there is the gathering of requirements.
2. Based on them, we write possible solutions into a document called a [Decision Record](#decision-record).
3. Then we have a discussion about the possible solutions.
4. When a solution is selected, we create a design document for it, describing the implementation steps.
5. And the last step is the implementation.

### Requirements

This step requires you to take your business expert (probably a Project Manager),
a whiteboard (it could be an online one), and bombard them with questions to figure out the problem.
Make sure you **understand** the problem well. Remember, there are no stupid questions.

Gather all functional and non-functional requirements and **write them down**. Without writing them down,
there is a high chance that something will be forgotten and will have to be implemented ad-hoc,
which introduces unnecessary tech debt into the system.

Now that you have a clear understanding of the problem, you can start looking for
solutions.

### Decision Record

> Also known as: Architecture Decision Records (ADRs), Architecture Description Records,
> Design System Decision Records, Lightweight Architecture Decision Records, Decision Log.

So what is it? In a more formal definition: A Decision Record is a record of software design choices that
addresses a functional or non-functional requirement that is **architecturally significant**. This
could be:
- a technology choice (e.g., PostgreSQL vs. MongoDB),
- an architectural decision (e.g., APIs design),
- a choice of library,
- a design of a feature (e.g., how to store the pictures of the cats).

In other words, this is a document where you keep your options on how to implement what needs to be done.
Any important decision should be captured in a structured way, meaning that the document's structure
is known, and it's easy to find the necessary information.
Here is [template](https://github.com/dmksnnk/blog/tree/main/examples/adr/000-template.md) we are using.

After you have your options, move to the review.

### Review

This is the phase where you decide which solution should be implemented. You have all your
ideas, present them to the world. Everyone on the team should read it to understand what's
going on and share their thoughts. If you are working on a solution that will affect other teams,
share it with them too. You may also ask people with more experience from other teams for a review.

Discuss and look for the best solution. There is no one-size-fits-all solution; you all have
a shared goal of finding the optimal solution given the requirements. It's all about
compromises and finding what suits you best.

After a consensus is reached, write down the selected solution. Now you can move on to the
design document.

### Design Document

Write down the concrete implementation details of what should be done:
- How the API will look.
- How the algorithm should work.
- What DB should be deployed.
- Tickets.
- etc.

That will also be your contract with other teams, so other teams can refer to it in their work.
It is OK if the actual implementation differs from what is described here. Update the document to reflect
reality, ideally with the reason why. It will be a great learning material.

### Implementation

The last step is implementation. If you have successfully struggled
through all the previous steps, this should be the easiest part:
you just need to translate what was decided into code.
And if you've done your job really well, you may leave it to an LLM to implement it.

---

Sources:
- [Architecture decision record (ADR)](https://github.com/joelparkerhenderson/architecture-decision-record)
- [Architectural decision](https://en.wikipedia.org/wiki/Architectural_decision)
- [Markdown Any Decision Records](https://adr.github.io/madr/)
- [Architecture Decisions: Demystifying Architecture](https://personal.utdallas.edu/~chung/SA/zz-Impreso-architecture_decisions-tyree-05.pdf)
- [How to create Architectural Decision Records (ADRs) — and how not to](https://www.ozimmer.ch/practices/2023/04/03/ADRCreation.html)
